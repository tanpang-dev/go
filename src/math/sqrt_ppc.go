// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// The 750CL (Wii) has no fsqrt; archSqrt builds a correctly rounded square
// root from frsqrte and fused multiply-add (see sqrt_ppc.s).
const haveArchSqrt = true

func archSqrt(x float64) float64
