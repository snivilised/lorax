package lo

// MIT License
//
// Copyright (c) 2022 Samuel Berthe

// Clonable defines a constraint of types having Clone() T method.
type Clonable[T any] interface {
	Clone() T
}
