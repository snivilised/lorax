package rx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"golang.org/x/sync/errgroup"
)

// This test file has been defined inside of rx as opposed to rx_test,
// because the tests depend on internal functionality.

var (
	errFoo = errors.New("foo")
	errBar = errors.New("bar")
)

// connect should go into factory-connectable_test
func collect[T any](ctx context.Context, ch <-chan Item[T]) ([]any, error) {
	s := make([]any, 0)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case item, ok := <-ch:
			if !ok {
				return s, nil
			}

			if item.IsError() {
				s = append(s, item.E)
			} else {
				s = append(s, item.V)
			}
		}
	}
}

var _ = Describe("FactoryConnectable", func() {
	Context("Connectable", func() {
		When("foo", func() {
			XIt("🧪 should: ", func() {
				defer leaktest.Check(GinkgoT())()
				// TODO(fix)
				// [FAILED] Expected
				// <*errors.errorString | 0xc000038090>:
				// expected: [1 2 3], got: []
				// {
				// 		s: "expected: [1 2 3], got: []",
				// }
				// to be nil
				ch := make(chan Item[int], 10)
				go func() {
					ch <- Of(1)
					ch <- Of(2)
					ch <- Of(3)
					close(ch)
				}()

				obs := &ObservableImpl[int]{
					iterable: newChannelIterable(ch, WithPublishStrategy[int]()),
				}

				testConnectableSingle(obs)
			})
		})
	})
})

func testConnectableSingle[T any](obs Observable[T]) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	eg, _ := errgroup.WithContext(ctx)
	expected := []interface{}{1, 2, 3}
	nbConsumers := 3
	wg := sync.WaitGroup{}

	wg.Add(nbConsumers)
	// Before Connect() is called we create multiple observers
	// We check all observers receive the same items
	for i := 0; i < nbConsumers; i++ {
		eg.Go(func() error {
			observer := obs.Observe(WithContext[T](ctx))

			wg.Done()

			got, err := collect(ctx, observer)

			if err != nil {
				return err
			}

			if !reflect.DeepEqual(got, expected) {
				return fmt.Errorf("expected: %v, got: %v", expected, got)
			}

			return nil
		})
	}

	wg.Wait()
	obs.Connect(ctx)

	Expect(eg.Wait()).Error().To(BeNil())
}

func testConnectableComposed[T any](obs Observable[T], increment Func[T]) {
	obs = obs.Map(func(ctx context.Context, v T) (T, error) {
		return increment(ctx, v) // expect client to implement with with v+1
	}, WithPublishStrategy[T]())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	const (
		nbConsumers = 3
	)

	eg, _ := errgroup.WithContext(ctx)
	expected := []interface{}{2, 3, 4} // TODO(check): is this correct?
	wg := sync.WaitGroup{}

	wg.Add(nbConsumers)
	// Before Connect() is called we create multiple observers
	// We check all observers receive the same items
	for i := 0; i < nbConsumers; i++ {
		eg.Go(func() error {
			observer := obs.Observe(WithContext[T](ctx))

			wg.Done()

			got, err := collect(ctx, observer)
			if err != nil {
				return err
			}

			if !reflect.DeepEqual(got, expected) {
				return fmt.Errorf("expected: %v, got: %v", expected, got)
			}

			return nil
		})
	}

	wg.Wait()
	obs.Connect(ctx)

	Expect(eg.Wait()).Error().To(BeNil())
}

func testConnectableWithoutConnect[T any](obs Observable[T]) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	Assert(ctx, obs, IsEmpty[T]())
}
