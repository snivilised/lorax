package rx

import (
	"fmt"

	"github.com/snivilised/lorax/enums"
)

// MIT License

// Copyright (c) 2016 Joe Chasinga

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// IllegalInputError is triggered when the observable receives an illegal input.
type IllegalInputError struct {
	error string
}

func (e IllegalInputError) Error() string {
	return "illegal input: " + e.error
}

// IndexOutOfBoundError is triggered when the observable cannot access to the specified index.
type IndexOutOfBoundError struct {
	error string
}

func (e IndexOutOfBoundError) Error() string {
	return "index out of bound: " + e.error
}

type TryError[T any] struct {
	error    string
	item     Item[T]
	expected enums.ItemDiscriminator
}

func (e TryError[T]) Error() string {
	return fmt.Sprintf("expected item to be : '%v', but is: '%v' (%+v)",
		enums.ItemDescriptions[e.expected], enums.ItemDescriptions[e.item.disc], e.item,
	)
}
