// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// func archSqrt(x float64) float64
//
// frsqrte gives about 5 bits of 1/sqrt(x); four Newton steps (with fused
// multiply-add) bring y to full precision. Then g = x*y and hh = y/2 are
// corrected Markstein-style: r = x - g*g (exact with fnmsub), g += r*hh,
// which yields the correctly rounded square root. Inputs are first scaled
// by an even power of two so that y*y stays a normal number.
TEXT ·archSqrt(SB),NOSPLIT,$0-16
	FMOVD	x+0(FP), F1
	FMOVD	$0.0, F0
	FCMPU	F1, F0
	BGT	pos
	BEQ	done	// ±0
	FSUB	F1, F1, F2	// negative or NaN: (x-x)/(x-x) is NaN
	FDIV	F2, F2, F1
	BR	done
pos:
	FMOVD	$0.5, F4
	FMUL	F1, F4, F2
	FCMPU	F2, F1	// only +Inf satisfies x/2 == x here
	BEQ	done

	FMOVD	$1.0, F9	// scale for the result
	FMOVD	$1.1830521861667747e-271, F2	// 2^-900
	FCMPU	F1, F2
	BGE	notiny
	FMOVD	$1.3407807929942597e+154, F2	// 2^512
	FMUL	F2, F1
	FMOVD	$8.636168555094445e-78, F9	// 2^-256
	BR	scaled
notiny:
	FMOVD	$1.0715086071862673e+301, F2	// 2^1000
	FCMPU	F1, F2
	BLE	scaled
	FMOVD	$7.458340731200207e-155, F2	// 2^-512
	FMUL	F2, F1
	FMOVD	$1.157920892373162e+77, F9	// 2^256
scaled:
	FRSQRTE	F1, F3	// y ≈ 1/sqrt(x)
	FMUL	F1, F4, F5	// h = x/2

	// y += y*(1/2 - h*y*y), four times
	FMUL	F3, F3, F6
	FNMSUB	F5, F4, F6, F6
	FMADD	F3, F3, F6, F3
	FMUL	F3, F3, F6
	FNMSUB	F5, F4, F6, F6
	FMADD	F3, F3, F6, F3
	FMUL	F3, F3, F6
	FNMSUB	F5, F4, F6, F6
	FMADD	F3, F3, F6, F3
	FMUL	F3, F3, F6
	FNMSUB	F5, F4, F6, F6
	FMADD	F3, F3, F6, F3

	FMUL	F1, F3, F7	// g = x*y ≈ sqrt(x)
	FMUL	F3, F4, F8	// hh = y/2 ≈ 1/(2 sqrt(x))

	FNMSUB	F7, F1, F7, F6	// r = x - g*g
	FMADD	F6, F7, F8, F7	// g += r*hh
	FNMSUB	F7, F4, F8, F6	// e = 1/2 - g*hh
	FMADD	F8, F8, F6, F8	// hh += hh*e
	FNMSUB	F7, F1, F7, F6	// r = x - g*g
	FMADD	F6, F7, F8, F7	// g += r*hh

	FMUL	F7, F9, F1	// undo the scaling (exact)
done:
	FMOVD	F1, ret+8(FP)
	RET
