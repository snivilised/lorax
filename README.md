# 🌟 lorax: ___Go Concurrency with Functional/Reactive Extensions___

[![A B](https://img.shields.io/badge/branching-commonflow-informational?style=flat)](https://commonflow.org)
[![A B](https://img.shields.io/badge/merge-rebase-informational?style=flat)](https://git-scm.com/book/en/v2/Git-Branching-Rebasing)
[![A B](https://img.shields.io/badge/branch%20history-linear-blue?style=flat)](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/managing-a-branch-protection-rule)
[![Go Reference](https://pkg.go.dev/badge/github.com/snivilised/lorax.svg)](https://pkg.go.dev/github.com/snivilised/lorax)
[![Go report](https://goreportcard.com/badge/github.com/snivilised/lorax)](https://goreportcard.com/report/github.com/snivilised/lorax)
[![Coverage Status](https://coveralls.io/repos/github/snivilised/lorax/badge.svg?branch=master)](https://coveralls.io/github/snivilised/lorax?branch=master&kill_cache=1)
[![lorax Continuous Integration](https://github.com/snivilised/lorax/actions/workflows/ci-workflow.yml/badge.svg)](https://github.com/snivilised/lorax/actions/workflows/ci-workflow.yml)
[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit&logoColor=white)](https://github.com/pre-commit/pre-commit)
[![A B](https://img.shields.io/badge/commit-conventional-commits?style=flat)](https://www.conventionalcommits.org/)

<!-- MD013/Line Length -->
<!-- MarkDownLint-disable MD013 -->

<!-- MD014/commands-show-output: Dollar signs used before commands without showing output mark down lint -->
<!-- MarkDownLint-disable MD014 -->

<!-- MD033/no-inline-html: Inline HTML -->
<!-- MarkDownLint-disable MD033 -->

<!-- MD040/fenced-code-language: Fenced code blocks should have a language specified -->
<!-- MarkDownLint-disable MD040 -->

<!-- MD028/no-blanks-blockquote: Blank line inside blockquote -->
<!-- MarkDownLint-disable MD028 -->

<!-- MD010/no-hard-tabs: Hard tabs -->
<!-- MarkDownLint-disable MD010 -->

<p align="left">
  <a href="https://go.dev"><img src="resources/images/go-logo-light-blue.png" width="50" alt="go.dev" /></a>
</p>

## 🔰 Introduction

___Reactive extensions for go (limited set of functionality as required by snivilised)___

This project has been setup for the currency needs of __snivilised__ projects. Ordinarily, I would have used [___RxGo___](https://github.com/ReactiveX/RxGo) for reactive functionality, but it is currently not in a working state and seems to be abandoned. So ___lorax___ will aim to fill the gap, implementing the functionality required by __snivilised__ projects (in particular ___extendio___, which requires the ability to navigate a directory tree concurrently).

The intention is to write a version of RxGo with identical semantics as the original, having fixed some of the outstanding issues in RxGo reported by other users and possibly introduce new features. If and when version 3 of RxGo becomes available, then lorax/rx will be considered surplice to requirements and retired.

___lorax___ will also provide concurrent functionality via an abstraction that implements the __worker pool__ pattern.

## 📚 Usage

Please refer to the documentation already available on [___RxGo___](https://github.com/ReactiveX/RxGo). The documentation here focuses on the specific differences of usage between lorax/rx vs RxGo.

💤 tbd

## 🎀 Features

<p align="left">
  <a href="https://onsi.github.io/ginkgo/"><img src="https://onsi.github.io/ginkgo/images/ginkgo.png" width="100" alt="ginkgo"/></a>
  <a href="https://onsi.github.io/gomega/"><img src="https://onsi.github.io/gomega/images/gomega.png" width="100" alt="gomega"/></a>
</p>

+ unit testing with [Ginkgo](https://onsi.github.io/ginkgo/)/[Gomega](https://onsi.github.io/gomega/)
+ implemented with [🐍 Cobra](https://cobra.dev/) cli framework, assisted by [🐲 Cobrass](https://github.com/snivilised/cobrass)
+ i18n with [go-i18n](https://github.com/nicksnyder/go-i18n)
+ linting configuration and pre-commit hooks, (see: [linting-golang](https://freshman.tech/linting-golang/)).

## 🚀 Rx

In the absence of a version 3 of [___RxGo___](https://github.com/ReactiveX/RxGo), lorax provides an alternative implementation, ___rx___, that uses generics over reflection. Use of generics simplifies some issues but introduces complexity in other areas. For example, the ___AverageXXX___ (eg ___AverageFloat32___) rxgo operator contains 'overloads' for the different numeric types. This overloading is not required when generics are in play and in __rx__, all the ___AverageXXX___ variants are replaced by a single ___Average___ operator.

Conversely, now that the channels have been made type specific, it is now no longer possible to send any type of value through them. An object of type __T__ can be sent via an instantiation of the newly modified ___Item___, which is now a generic, based upon type __T__. But there are times when we need to send values which are not of type T. To keep using the same channel, a solution was required to resolve this issue. This solution involves the introduction of a new field on ___Item[T]___; ___opaque___ and a discriminator field ___disc___. The opaque field is defined as ___any___ and its type is identified by the ___disc___ member. There are additional helpers to create the ___Item___ instance similar to the existing ___Of___ function and also getter functions.

The following sections describe the issues encountered when implementing generics and how to use different aspects of the rx library using the new paradigm.

### 💎 Item

The new representation of an Item is as follows:

```go
Item[T any] struct {
  V      T
  E      error
  opaque any
  disc   enums.ItemDiscriminator
}
```

To minimise change, the ___V___ member remains in place and the semantics behind it remain the same. The opaque field is used to carry items not of type T. The presence of V is mutually exclusive with opaque and they are distinguished by the discriminator. An opaque instance of ___Item[T]___ is used whenever the type of value that needs to be send is independent of type ___T___. For example consider this __rx__ code:

```go
func (op *allOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if op.all {
		True[T]().SendContext(ctx, dst)
	}
}
```

The ___end___ method on ___allOperator___ needs to send a boolean true value (which is of course independent of type ___T___) through the channel. The ___True[T]___ helper function, creates an instance of ___Item[T]___ with the opaque field set to ___true___ and the discriminator set to ___enums.ItemDiscBoolean___.

#### 🎓 Item helpers

The receiver of an ___Item[T]___ needs to know what type of instance it is. Most of the time, the item will just carry a native value ___T___ and just because of the scenario in which the item is being received, the receiver knows that its a native instance. In these cases, the receiver can directly access the V member. However, there are times when the receiver does not know what type the item is.

Various try/must helpers have been defined to make this task easier for clients. Eg, consider this helper for boolean values:

```go
func TryBool[T any](item Item[T]) (bool, error)
```

This will return the opaque field cast as a bool, if the discriminator says its valid to do so or an error otherwise. There are other similar helpers for a variety of supported auxiliary types (see __must-try.go__). There are also ___MustXXX___ equivalents for each type that panic instead of returning an error.

### 🔢 Calculator

Various operators need to be able to compute simple numerical calculations on item values. This is no longer possible to do directly when the type of T is unknown. This has been resolved by the introduction of a calculator that must be specified as an option when using any operation that requires a numerical calculation.

The following is an illustration of using the new ___WithCalc___ option used in conjunction with the ___Assert___ api:

```go
  rx.Assert(ctx,
    testObservable[float32](ctx,
      float32(1), float32(20),
    ).Average(rx.WithCalc(rx.Calc[float32]())),
    rx.HasItem[float32]{
      Expected: 10.5,
    },
  )
```

Since the calc operation is common to a variety of functions/operators, to aid simplicity of use, its better to have to only specify this once, rather than pass in a calc directly into the function/operator in use; doing so in this case where multiple operators are in play, would require passing in a calc to each and every one. ___WithCalc___ allows us to have to only specify this once. Of course, with the calculator being an option, it is easier to forget to pass one in when it is required. In this case, the operation will fail and an error returned. Any operation requiring a calc is documented as such.

NB: rx.Calc is a pre-defined calculator instance that works for all numeric scalar types.

### 🌀 Iteration

The ___Range___ function needed special attention, because of the need to send numeric values through channels. Consider the legacy implementation of ___rangeIterable___:

```go
  for idx := i.start; idx <= i.start+i.count-1; idx++ {
    select {
    case <-ctx.Done():
      return
    case next <- Of(idx):
    }
  }
```

The index ___idx___ is an int and is sent as such, but this is no longer possible. We could have used item as a ___Num___ instance, but doing so would have caused rippling knock-on effects through a chain of operators. That is to say, when ___Range___ is used in combination with an operator like ___Max___, then Max would also have to be aware of ___Item___ being a ___Num___. This is undesirable, so instead, the iterator variable is declared to be of type T and therefore can be carried by a native ___Item___. This works ok for ___numeric___ scalar types (float32, int, ...), but for compound types this is a bit more complicated, imposing constraints on the type T (see ___ProxyField___).

We now have 2 Range functions, 1 for numeric scalar types and the other for compound struct types. The body of both of these are the same:

```go
  for idx, _ := i.iterator.Start(); i.iterator.While(*idx); i.iterator.Increment(idx) {
    select {
    case <-ctx.Done():
      return
    case next <- Of(*idx):
    }
  }
```

The iterator contains methods, that should be implemented by the client. ___Start___ creates an initial value of type T and returns a pointer to it. ___While___ is a condition function that returns true whilst the iteration should continue to run and ___Increment___ receives a pointer to the index variable so that it can be incremented.

The following is an example of how to use the Range operator for scalar types:

```go
  obs := rx.Range(&rx.NumericRangeIterator[int]{
    StartAt: 5,
    By: 1,
    Whilst:  rx.LessThan(8),
  })
```

___NumericRangeIterator___ is defined for all numeric types and is therefore able to use the native operators for calculation operations.

For struct types, the above is identical, but instead using ___ProxyRangeIterator___:

```go
  obs := rx.RangePF(&rx.ProxyRangeIterator[widget, int]{
    StartAt: widget{id: 5},
    By:      widget{id: 1},
    Whilst:  rx.LessThanPF(widget{id: 8}),
  })
```

The user can also perform reverse iteration by using a negative ___By___ value and then using the pre-0defined ___rx.MoreThanPF___ for the ___Whilst___ condition function.

#### 🎙️ Proxy field

In order for ___RangePF___ to work on struct types, a constraint has to be placed on type T. We need type T to have certain methods and is defined as:

```go
	Numeric interface {
		constraints.Integer | constraints.Signed | constraints.Unsigned | constraints.Float
	}

	ProxyField[T any, O Numeric] interface {
		Field() O
		Inc(index *T, by T) *T
		Index(i int) *T
	}
```

So for a type T, T has to nominate a member field (the proxy) which will act as the iterator value of type O and this is the purpose of the ___Field___ function. ___Inc___ is the function that performs the increment, by the value specified as ___By___ and ___Index___ allows us to derive an index value of type T from an integer source.

This makes for a flexible approach that allows the type T to be in control of incrementing the index value over each iteration.

So given a domain type widget:

```go
type widget struct {
	id     int
	name   string
	amount int
}
```

we nominate ___id___ as the proxy field:

```go
func (w widget) Field() int {
	return w.id
}
```

Our incrementor is defined as:

```go
func (w widget) Inc(index *widget, by widget) *widget {
	index.id += by.id

	return index
}
```

This may look strange, but is necessary since the type T can not be defined with pointer receivers with respect to the ___Numeric___ constraint. The reason for this is to keep in line with the original rxgo functionality of being able to compose an observable with literal scalar values and we can't take the address of literal scalars that would be required in order to be able to define ___Inc___ as:

> func (w *widget) Inc(by widget) *widget

So ___Numeric___ receivers on T being of the non pointer variety is a strict invariant.

The aspect to focus on in ___widget.Inc___ is that ___index___ is incremented with the ___by___ value, not ___w.id___. Effectively, widget is passed a pointer to its original self as index, but w is the copy of index in which we're are running. For this to work properly, the original widget (index) must be incremented, not the copy (w), which would have no effect, resulting in an infinite loop owing to the exit condition never being met.

### 🎭 Map

The ___Map___ functionality poses a new challenge. If we wanted to map a value of type T to a value other than T, the mapped to value could not be sent through the channel because it is of the wrong type. A work-around would be to use an opaque instance of Item, but then that could easily become very messy as we no no longer have consistent types of emitted values which would be difficult to keep track of.

So, in the short term, what we say is that the ___Map___ operator works, but can only map to different values of the same type, but this is also a little too restricting. Mapping to a different type is not a niche feature, but we can speculate as to what a solution would look like (this is not implemented yet, to be addressed under issue [#230](https://github.com/snivilised/lorax/issues/230))

The essence of the problem is that we need to represent the new type O required for ___Map___. But we can't introduce O to Item, as that would mean every other generic type would also need to be aware of O. This is incorrect as the type O would only be needed for Map and nothing else. There is no such facility to be able to define a void type.

To fix this, we need some kind of bridging mechanism. This could be either a function or a struct, which would be defined to accept this second generic parameter. The bridge connects a source observable based on type T to a new observable chain based on type O.

## 😵‍💫 Trouble shooting

### Infinite range iteration

Be careful how the range iterator is specified. Make sure that the incrementor defined by ___By___ tends towards the exit condition specified by ___Whilst___, otherwise infinity will ensue.

### Missing calc

Make sure than any operator that needs a calculator is provided one via the ___WithCalc___ option. A missing calc will result in an error that indicates as such.
