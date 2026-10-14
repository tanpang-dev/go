// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !ppc

package math

const haveArchSqrt = false

func archSqrt(x float64) float64 {
	panic("not implemented")
}
