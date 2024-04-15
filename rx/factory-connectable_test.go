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

func increment(_ context.Context, v int) (int, error) {
	return v + 1, nil
}

var _ = Describe("FactoryConnectable", func() {
	Context("Connectable iterable channel", func() {
		When("single with multiple observers", func() {
			It("🧪 should: distribute all items to all connected observers", func() {
				// Test_Connectable_IterableChannel_Single
				defer leaktest.Check(GinkgoT())()

				ch := make(chan Item[int], 10)
				go func() {
					ch <- Of(1)
					ch <- Of(2)
					ch <- Of(3)
					close(ch)
				}()

				obs := &ObservableImpl[int]{
					iterable: newChannelIterable(ch,
						WithPublishStrategy[int](),
					),
				}

				testConnectableSingle(obs, []any{1, 2, 3})
			})
		})
	})

	Context("Connectable iterable channel", func() {
		When("composed with multiple observers", func() {
			It("🧪 should: distribute all items to all connected observers", func() {
				// Test_Connectable_IterableChannel_Composed
				defer leaktest.Check(GinkgoT())()

				ch := make(chan Item[int], 10)
				go func() {
					ch <- Of(1)
					ch <- Of(2)
					ch <- Of(3)
					close(ch)
				}()

				obs := &ObservableImpl[int]{
					iterable: newChannelIterable(ch,
						WithPublishStrategy[int](),
					),
				}

				testConnectableComposed(obs, increment, []any{2, 3, 4})
			})
		})
	})

	Context("Connectable iterable channel", func() {
		When("disposed", func() {
			It("🧪 should: distribute all items to all connected observers", func() {
				// Test_Connectable_IterableChannel_Disposed
				defer leaktest.Check(GinkgoT())()

				ch := make(chan Item[int], 10)
				go func() {
					ch <- Of(1)
					ch <- Of(2)
					ch <- Of(3)
					close(ch)
				}()

				obs := &ObservableImpl[int]{
					iterable: newChannelIterable(ch,
						WithPublishStrategy[int](),
					),
				}

				ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
				defer cancel()

				_, disposable := obs.Connect(ctx)
				disposable()
				time.Sleep(50 * time.Millisecond)
				Assert(ctx, obs, IsEmpty[int]{})
			})
		})
	})

	Context("Connectable iterable channel", func() {
		When("Without connect", func() {
			It("🧪 should: distribute all items to all connected observers", func() {
				// Test_Connectable_IterableChannel_WithoutConnect
				defer leaktest.Check(GinkgoT())()

				ch := make(chan Item[int], 10)
				go func() {
					ch <- Of(1)
					ch <- Of(2)
					ch <- Of(3)
					close(ch)
				}()

				obs := &ObservableImpl[int]{
					iterable: newChannelIterable(ch,
						WithPublishStrategy[int](),
					),
				}
				testConnectableWithoutConnect(obs)
			})
		})
	})

	Context("Connectable iterable", func() {
		When("create single", func() {
			It("🧪 should: distribute all items to all connected observers", func() {
				// Test_Connectable_IterableCreate_Single
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := &ObservableImpl[int]{
					iterable: newCreateIterable([]Producer[int]{func(_ context.Context, ch chan<- Item[int]) {
						ch <- Of(1)
						ch <- Of(2)
						ch <- Of(3)
						cancel()
					}},
						WithPublishStrategy[int](),
						WithContext[int](ctx),
					),
				}
				testConnectableSingle(obs, []any{1, 2, 3})
			})
		})
	})

	Context("Connectable iterable", func() {
		When("composed", func() {
			It("🧪 should: distribute all items to all connected observers", func() {
				// Test_Connectable_IterableCreate_Composed
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := &ObservableImpl[int]{
					iterable: newCreateIterable([]Producer[int]{func(_ context.Context, ch chan<- Item[int]) {
						ch <- Of(1)
						ch <- Of(2)
						ch <- Of(3)
						cancel()
					}},
						WithPublishStrategy[int](),
						WithContext[int](ctx),
					),
				}
				testConnectableComposed(obs, increment, []any{2, 3, 4})
			})
		})
	})

	Context("Connectable iterable create", func() {
		When("disposed", func() {
			It("🧪 should: distribute all items to all connected observers", func() {
				// Test_Connectable_IterableCreate_Disposed
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := &ObservableImpl[int]{
					iterable: newCreateIterable([]Producer[int]{func(_ context.Context, ch chan<- Item[int]) {
						ch <- Of(1)
						ch <- Of(2)
						ch <- Of(3)
						cancel()
					}},
						WithPublishStrategy[int](),
						WithContext[int](ctx),
					),
				}
				obs.Connect(ctx)
				_, cancel2 := context.WithTimeout(context.Background(), 550*time.Millisecond)
				defer cancel2()
				time.Sleep(50 * time.Millisecond)
				Assert(ctx, obs, IsEmpty[int]{})
			})
		})

		When("Without connect", func() {
			It("🧪 should: distribute all items to all connected observers", func() {
				// Test_Connectable_IterableCreate_WithoutConnect
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := &ObservableImpl[int]{
					iterable: newCreateIterable([]Producer[int]{func(_ context.Context, ch chan<- Item[int]) {
						ch <- Of(1)
						ch <- Of(2)
						ch <- Of(3)
						cancel()
					}}, WithBufferedChannel[int](3),
						WithPublishStrategy[int](),
						WithContext[int](ctx),
					),
				}
				testConnectableWithoutConnect(obs)
			})
		})

		Context("Connectable iterable", func() {
			When("defer single", func() {
				It("🧪 should: distribute all items to all connected observers", func() {
					// Test_Connectable_IterableDefer_Single
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := &ObservableImpl[int]{
						iterable: newDeferIterable([]Producer[int]{func(_ context.Context, ch chan<- Item[int]) {
							ch <- Of(1)
							ch <- Of(2)
							ch <- Of(3)
							cancel()
						}},
							WithBufferedChannel[int](3),
							WithPublishStrategy[int](),
							WithContext[int](ctx),
						),
					}
					testConnectableSingle(obs, []any{1, 2, 3})
				})
			})

			When("defer composed", func() {
				It("🧪 should: distribute all items to all connected observers", func() {
					// Test_Connectable_IterableDefer_Composed
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := &ObservableImpl[int]{
						iterable: newDeferIterable([]Producer[int]{func(_ context.Context, ch chan<- Item[int]) {
							ch <- Of(1)
							ch <- Of(2)
							ch <- Of(3)
							cancel()
						}},
							WithBufferedChannel[int](3),
							WithPublishStrategy[int](),
							WithContext[int](ctx),
						),
					}
					testConnectableComposed(obs, increment, []any{2, 3, 4})
				})
			})

			When("Just single", func() {
				It("🧪 should: distribute all items to all connected observers", func() {
					// Test_Connectable_IterableJust_Single
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := &ObservableImpl[int]{
						iterable: newJustIterable[int](1, 2, 3)(
							WithPublishStrategy[int](),
							WithContext[int](ctx),
						),
					}
					testConnectableSingle(obs, []any{1, 2, 3})
				})
			})

			When("Just composed", func() {
				It("🧪 should: distribute all items to all connected observers", func() {
					// Test_Connectable_IterableJust_Composed
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := &ObservableImpl[int]{
						iterable: newJustIterable[int](1, 2, 3)(
							WithPublishStrategy[int](),
							WithContext[int](ctx),
						),
					}
					testConnectableComposed(obs, increment, []any{2, 3, 4})
				})
			})

			When("range single", func() {
				It("🧪 should: distribute all items to all connected observers", func() {
					// Test_Connectable_IterableRange_Single
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := &ObservableImpl[int]{
						iterable: newRangeIterable(1, 3,
							WithPublishStrategy[int](),
							WithContext[int](ctx),
						),
					}
					testConnectableSingle(obs, []any{1, 2, 3})
				})
			})
		})
	})
})

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

			switch {
			case item.IsError():
				s = append(s, item.E)
			case item.IsNumber():
				s = append(s, item.Num())
			default:
				s = append(s, item.V)
			}
		}
	}
}

func testConnectableSingle[T any](obs Observable[T], expected []any) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	const (
		nbConsumers = 3
	)

	eg, _ := errgroup.WithContext(ctx)
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

	Expect(eg.Wait()).To(Succeed())
}

func testConnectableComposed[T any](obs Observable[T], increment Func[T], expected []any) {
	obs = obs.Map(func(ctx context.Context, v T) (T, error) {
		return increment(ctx, v) // expect client to implement with v+1
	}, WithPublishStrategy[T]())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	const (
		nbConsumers = 3
	)

	eg, _ := errgroup.WithContext(ctx)
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

	Assert(ctx, obs, IsEmpty[T]{})
}
