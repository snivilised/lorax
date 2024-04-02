package rx_test

import (
	"context"
	"errors"
	"time"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/samber/lo"

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Factory", func() {
	Context("Amb", func() {
		When("Amb1??", func() {
			It("🧪 should: emit from first responding observer only", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Amb([]rx.Observable[int]{
					testObservable[int](ctx, 1, 2, 3),
					rx.Empty[int](),
				})
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}))
			})
		})

		When("Amb2??", func() {
			It("🧪 should: emit from first responding observer only", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Amb([]rx.Observable[int]{
					rx.Empty[int](),
					testObservable[int](ctx, 1, 2, 3),
					rx.Empty[int](),
					rx.Empty[int](),
				})
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}))
			})
		})
	})

	Context("CombineLatest", func() {
		When("Multiple observables", func() {
			It("🧪 should: combine", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.CombineLatest(func(values ...int) int {
					return lo.Sum(values)
				}, lo.Map([]rx.Observable[int]{
					testObservable[int](ctx, 1, 2),
					testObservable[int](ctx, 10, 11),
				}, func(it rx.Observable[int], _ int) rx.Observable[int] {
					return it
				}))

				rx.Assert(context.Background(), obs, rx.IsNotEmpty[int]())
			})
		})

		When("Empty", func() {
			It("🧪 should: be able to detect empty observable", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.CombineLatest(func(values ...int) int {
					return lo.Sum(values)
				}, lo.Map([]rx.Observable[int]{
					testObservable[int](ctx, 1, 2),
					rx.Empty[int](),
				}, func(it rx.Observable[int], _ int) rx.Observable[int] {
					return it
				}))

				rx.Assert(context.Background(), obs, rx.IsEmpty[int]())
			})
		})

		When("Contains error", func() {
			It("🧪 should: be able to detect error", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.CombineLatest(func(values ...int) int {
					return lo.Sum(values)
				}, lo.Map([]rx.Observable[int]{
					testObservable[int](ctx, 1, 2),
					testObservable[int](ctx, errFoo),
				}, func(it rx.Observable[int], _ int) rx.Observable[int] {
					return it
				}))

				rx.Assert(context.Background(), obs, rx.IsEmpty[int](), rx.HasError[int](errFoo))
			})
		})
	})

	Context("Concat", func() {
		When("Single observable", func() {
			It("🧪 should: create derived single observable", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Concat([]rx.Observable[int]{
					testObservable[int](ctx, 1, 2, 3),
				})
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}))
			})
		})

		When("Two observables", func() {
			It("🧪 should: create derived compound single observable", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Concat([]rx.Observable[int]{
					testObservable[int](ctx, 1, 2, 3),
					testObservable[int](ctx, 4, 5, 6),
				})
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3, 4, 5, 6}))
			})
		})

		When("More than two observables", func() {
			It("🧪 should: create derived compound single observable", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Concat([]rx.Observable[int]{
					testObservable[int](ctx, 1, 2, 3),
					testObservable[int](ctx, 4, 5, 6),
					testObservable[int](ctx, 7, 8, 9),
				})
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3, 4, 5, 6, 7, 8, 9}))
			})
		})

		When("Multiple empty observables", func() {
			It("🧪 should: create derived compound single observable", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Concat([]rx.Observable[int]{
					rx.Empty[int](),
					rx.Empty[int](),
					rx.Empty[int](),
				})
				rx.Assert(context.Background(), obs, rx.IsEmpty[int]())
			})
		})

		When("One empty observable", func() {
			It("🧪 should: create derived compound single observable", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Concat([]rx.Observable[int]{
					rx.Empty[int](),
					testObservable[int](ctx, 1, 2, 3),
				})
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}))

				obs = rx.Concat([]rx.Observable[int]{
					testObservable[int](ctx, 1, 2, 3),
					rx.Empty[int](),
				})
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}))
			})
		})
	})

	Context("Create", func() {
		When("provided with a Producer", func() {
			It("🧪 should: create observable", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Create([]rx.Producer[int]{func(_ context.Context, next chan<- rx.Item[int]) {
					next <- rx.Of(1)
					next <- rx.Of(2)
					next <- rx.Of(3)
				}})
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}), rx.HasNoError[int]())
			})
		})

		When("Provided with a Producer", func() {
			It("🧪 should: create observable (single dup?)", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Create([]rx.Producer[int]{func(_ context.Context, next chan<- rx.Item[int]) {
					next <- rx.Of(1)
					next <- rx.Of(2)
					next <- rx.Of(3)
				}})
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}), rx.HasNoError[int]())
				rx.Assert(context.Background(), obs, rx.IsEmpty[int](), rx.HasNoError[int]())
			})
		})

		When("context cancelled", func() {
			It("🧪 should: create observable", func() {
				defer leaktest.Check(GinkgoT())()

				closed1 := make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())

				_ = rx.Create([]rx.Producer[int]{
					func(_ context.Context, _ chan<- rx.Item[int]) {
						cancel()
					},
					func(ctx context.Context, _ chan<- rx.Item[int]) {
						<-ctx.Done()
						closed1 <- struct{}{}
					},
				}, rx.WithContext[int](ctx)).Run()

				select {
				case <-time.Tick(time.Second):
					Fail("producer not closed")

				case <-closed1:
				}
			})
		})
	})

	Context("Defer", func() {
		When("single", func() {
			It("🧪 should: create deferred observer", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Defer([]rx.Producer[int]{
					func(_ context.Context, next chan<- rx.Item[int]) {
						next <- rx.Of(1)
						next <- rx.Of(2)
						next <- rx.Of(3)
					}})
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}), rx.HasNoError[int]())
			})
		})

		When("multiple", func() {
			It("should: create deferred observer", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Defer([]rx.Producer[int]{
					func(_ context.Context, next chan<- rx.Item[int]) {
						next <- rx.Of(1)
						next <- rx.Of(2)
					},
					func(_ context.Context, next chan<- rx.Item[int]) {
						next <- rx.Of(10)
						next <- rx.Of(20)
					},
				})

				rx.Assert(context.Background(), obs, rx.HasItemsNoOrder(1, 2, 10, 20), rx.HasNoError[int]())
			})
		})

		When("context cancelled", func() {
			It("🧪 should: create deferred observable", func() {
				defer leaktest.Check(GinkgoT())()

				closed1 := make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())

				_ = rx.Defer([]rx.Producer[int]{
					func(_ context.Context, _ chan<- rx.Item[int]) {
						cancel()
					},
					func(ctx context.Context, _ chan<- rx.Item[int]) {
						<-ctx.Done()
						closed1 <- struct{}{}
					},
				}, rx.WithContext[int](ctx)).Run()

				select {
				case <-time.Tick(time.Second):
					Fail("producer not closed")

				case <-closed1:
				}
			})
		})

		When("Provided with a Producer", func() {
			It("🧪 should: create deferred observable (single dup?)", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Defer([]rx.Producer[int]{func(_ context.Context, next chan<- rx.Item[int]) {
					next <- rx.Of(1)
					next <- rx.Of(2)
					next <- rx.Of(3)
				}})
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}), rx.HasNoError[int]())
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}), rx.HasNoError[int]())
			})
		})

		When("ComposeDup", func() {
			It("🧪 should: create deferred observable (composed dup?)", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Defer([]rx.Producer[int]{func(_ context.Context, next chan<- rx.Item[int]) {
					next <- rx.Num[int](1)
					next <- rx.Num[int](2)
					next <- rx.Num[int](3)
				}}).Map(func(_ context.Context, i int) (_ int, _ error) {
					return i + 1, nil
				}).Map(func(_ context.Context, i int) (_ int, _ error) {
					return i + 1, nil
				})
				rx.Assert(context.Background(), obs, rx.HasNumbers[int]([]int{3, 4, 5}), rx.HasNoError[int]())
				rx.Assert(context.Background(), obs, rx.HasNumbers[int]([]int{3, 4, 5}), rx.HasNoError[int]())
			})
		})

		When("ComposeDup with eager observation", func() {
			It("🧪 should: create deferred observable (composed dup?)", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Defer([]rx.Producer[int]{func(_ context.Context, next chan<- rx.Item[int]) {
					next <- rx.Num[int](1)
					next <- rx.Num[int](2)
					next <- rx.Num[int](3)
				}}).Map(func(_ context.Context, i int) (_ int, _ error) {
					return i + 1, nil
				}, rx.WithObservationStrategy[int](rx.Eager)).Map(func(_ context.Context, i int) (_ int, _ error) {
					return i + 1, nil
				})
				rx.Assert(context.Background(), obs, rx.HasNumbers[int]([]int{3, 4, 5}), rx.HasNoError[int]())
				// In the case of an eager observation, we already consumed the items produced by Defer
				// So if we create another subscription, it will be empty
				rx.Assert(context.Background(), obs, rx.IsEmpty[int](), rx.HasNoError[int]())
			})
		})

		When("Error", func() {
			It("🧪 should: be detectable in observable", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Defer([]rx.Producer[int]{func(_ context.Context, next chan<- rx.Item[int]) {
					next <- rx.Of(1)
					next <- rx.Of(2)
					next <- rx.Error[int](errFoo)
				}})
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2}), rx.HasError[int](errFoo))
			})
		})
	})

	Context("Empty", func() {
		It("🧪 should: contain no elements", func() {
			defer leaktest.Check(GinkgoT())()

			obs := rx.Empty[int]()
			rx.Assert(context.Background(), obs, rx.IsEmpty[int]())
		})
	})

	Context("FromChannel", func() {
		It("🧪 should: create observable from channel", func() {
			defer leaktest.Check(GinkgoT())()

			ch := make(chan rx.Item[int])
			go func() {
				ch <- rx.Of(1)
				ch <- rx.Of(2)
				ch <- rx.Of(3)
				close(ch)
			}()

			obs := rx.FromChannel(ch)
			rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}), rx.HasNoError[int]())
		})

		When("SimpleCapacity", func() {
			It("🧪 should: ???", func() {
				defer leaktest.Check(GinkgoT())()

				ch := rx.FromChannel(make(chan rx.Item[int], 10)).Observe()
				Expect(cap(ch)).To(Equal(10))
			})
		})

		When("ComposedCapacity", func() {
			XIt("🧪 should: ???", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				obs1 := rx.FromChannel(make(chan rx.Item[int], 10)).
					Map(func(_ context.Context, _ int) (int, error) {
						return 1, nil
					}, rx.WithContext[int](ctx), rx.WithBufferedChannel[int](11))

				Expect(cap(obs1.Observe())).To(Equal(11))

				obs2 := obs1.Map(func(_ context.Context, _ int) (int, error) {
					return 1, nil
				}, rx.WithContext[int](ctx), rx.WithBufferedChannel[int](12))

				// FAILED => Observe returns 0
				Expect(cap(obs2.Observe())).To(Equal(12))
			})
		})
	})

	Context("FromEventSource", func() {
		When("Observation after all sent", func() {
			It("🧪 should: not see any items", func() {
				defer leaktest.Check(GinkgoT())()

				const max = 10
				next := make(chan rx.Item[int], max)
				obs := rx.FromEventSource(next, rx.WithBackPressureStrategy[int](rx.Drop))

				go func() {
					for i := 0; i < max; i++ {
						next <- rx.Of(i)
					}
					close(next)
				}()
				time.Sleep(50 * time.Millisecond)

				rx.Assert(context.Background(), obs, rx.CustomPredicate(func(items []int) error {
					if len(items) != 0 {
						return errors.New("items should be nil")
					}

					return nil
				}))
			})
		})

		When("Drop", func() {
			It("🧪 should: ???", func() {
				defer leaktest.Check(GinkgoT())()

				const max = 100000
				next := make(chan rx.Item[int], max)
				obs := rx.FromEventSource(next, rx.WithBackPressureStrategy[int](rx.Drop))

				go func() {
					for i := 0; i < max; i++ {
						next <- rx.Of(i)
					}
					close(next)
				}()

				rx.Assert(context.Background(), obs, rx.CustomPredicate(func(items []int) error {
					if len(items) == max {
						return errors.New("some items should be dropped")
					}
					if len(items) == 0 {
						return errors.New("no items")
					}

					return nil
				}))
			})
		})
	})

	Context("JustItem", func() {
		When("given: a value", func() {
			It("🧪 should: return a single item observable containing value", func() {
				defer leaktest.Check(GinkgoT())()

				single := rx.JustItem(1)
				rx.Assert(context.Background(), single, rx.HasItem(1), rx.HasNoError[int]())
				rx.Assert(context.Background(), single, rx.HasItem(1), rx.HasNoError[int]())
			})
		})
	})

	Context("Just", func() {
		When("given: a value", func() {
			It("🧪 should: return a single item observable containing value", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Just(1, 2, 3)()
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}), rx.HasNoError[int]())
				rx.Assert(context.Background(), obs, rx.HasItems([]int{1, 2, 3}), rx.HasNoError[int]())
			})
		})

		When("given: custom structure", func() {
			It("🧪 should:  ", func() {
				defer leaktest.Check(GinkgoT())()

				type customer struct {
					id int
				}

				obs := rx.Just([]customer{{id: 1}, {id: 2}, {id: 3}}...)()
				rx.Assert(context.Background(), obs,
					rx.HasItems([]customer{{id: 1}, {id: 2}, {id: 3}}),
					rx.HasNoError[customer](),
				)

				rx.Assert(context.Background(), obs,
					rx.HasItems([]customer{{id: 1}, {id: 2}, {id: 3}}),
					rx.HasNoError[customer](),
				)
			})
		})

		When("given: channel", func() {
			XIt("🧪 should: ???", func() {
				defer leaktest.Check(GinkgoT())()

				ch := make(chan int, 1)
				go func() {
					ch <- 1
					ch <- 2
					ch <- 3
					close(ch)
				}()
				obs := rx.Just(ch)()
				_ = obs

				// TODO(fix): o := rx.Ch[int](1)
				// rx.Assert(context.Background(), obs, rx.HasItems[int]([]int{1, 2, 3}))
			})
		})

		When("given: simple capacity", func() {
			It("🧪 should: ???", func() {
				defer leaktest.Check(GinkgoT())()

				ch := rx.Just(1)(rx.WithBufferedChannel[int](5)).Observe()
				Expect(cap(ch)).To(Equal(5))
			})
		})

		When("given: composed capacity", func() {
			XIt("🧪 should: ???", func() {
				defer leaktest.Check(GinkgoT())()

				obs1 := rx.Just(1)().Map(func(_ context.Context, _ int) (int, error) {
					return 1, nil
				}, rx.WithBufferedChannel[int](11))
				// FAILED => Observe returns 0
				Expect(cap(obs1.Observe())).To(Equal(11))

				obs2 := obs1.Map(func(_ context.Context, _ int) (int, error) {
					return 1, nil
				}, rx.WithBufferedChannel[int](12))
				Expect(cap(obs2.Observe())).To(Equal(12))
			})
		})
	})

	Context("Merge", func() {
		When("given, multiple observers", func() {
			It("🧪 should: combine into a single observer", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Merge([]rx.Observable[int]{
					testObservable[int](ctx, 1, 2),
					testObservable[int](ctx, 3, 4),
				})
				rx.Assert(context.Background(), obs, rx.HasItemsNoOrder(1, 2, 3, 4))
			})
		})

		When("given, multiple observers and contains error", func() {
			It("🧪 should: able to detect error in combined observable", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Merge([]rx.Observable[int]{
					testObservable[int](ctx, 1, 2),
					testObservable[int](ctx, 3, errFoo),
				})

				// The content is not deterministic, hence we just test if we have some items
				rx.Assert(context.Background(), obs, rx.IsNotEmpty[int](), rx.HasError[int](errFoo))
			})
		})
	})

	Context("Range", func() {
		When("positive count", func() {
			XIt("🧪 should: create observable", func() {
				defer leaktest.Check(GinkgoT())()

				/*
					this code inside rx.Assert:
					if item.IsError() {
						errs = append(errs, item.E)
					} else {
						got = append(got, item.V)
					}
					needs to accommodate item.N, ie the numeric aux value
					and also should be modified to support all the other
					new ways of interpreting an item (Ch, Tick, Tv)
				*/

				const (
					start = 5
					count = 3
				)

				obs := rx.Range[int](start, count)
				rx.Assert(context.Background(), obs, rx.HasItems([]int{5, 6, 7}))
				// Test whether the observable is reproducible
				rx.Assert(context.Background(), obs, rx.HasItems([]int{5, 6, 7}))
			})
		})

		When("negative count", func() {
			It("🧪 should: contain detectable error", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Range[int](1, -5)
				rx.Assert(context.Background(), obs, rx.HasAnError[int]())
			})
		})

		When("maximum exceeded", func() {
			It("🧪 should: contain detectable error", func() {
				defer leaktest.Check(GinkgoT())()

				const (
					start = 1 << 31
					count = 1
				)

				obs := rx.Range[int](start, count)
				rx.Assert(context.Background(), obs, rx.HasAnError[int]())
			})
		})
	})

	Context("Start", func() {
		When("using Supplier", func() {
			It("🧪 should: ???", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Start([]rx.Supplier[int]{func(_ context.Context) rx.Item[int] {
					return rx.Of(1)
				}, func(_ context.Context) rx.Item[int] {
					return rx.Of(2)
				}})
				rx.Assert(context.Background(), obs, rx.HasItemsNoOrder(1, 2))
			})
		})
	})

	Context("Thrown", func() {
		When("foo", func() {
			It("🧪 should: ", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Thrown[int](errFoo)
				rx.Assert(context.Background(), obs, rx.HasError[int](errFoo))
			})
		})
	})

	Context("Timer", func() {
		When("foo", func() {
			It("🧪 should: ???", func() {
				defer leaktest.Check(GinkgoT())()

				obs := rx.Timer[int](rx.WithDuration(time.Nanosecond))
				select {
				case <-time.Tick(time.Second):
					Fail("observable not closed")
				case <-obs.Observe():
				}
			})
		})

		When("Empty", func() {
			It("🧪 should: ???", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				obs := rx.Timer(rx.WithDuration(time.Hour), rx.WithContext[int](ctx))

				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()

				select {
				case <-time.Tick(time.Second):
					Fail("observable not closed")
				case <-obs.Observe():
				}
			})
		})
	})
})