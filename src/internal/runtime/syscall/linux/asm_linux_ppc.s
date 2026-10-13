// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && ppc

#include "textflag.h"

// func Syscall6(num, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, errno uintptr)
TEXT ·Syscall6(SB),NOSPLIT|NOFRAME,$0-40
	MOVW	num+0(FP), R31	// syscall entry
	MOVW	a1+4(FP), R3
	MOVW	a2+8(FP), R4
	MOVW	a3+12(FP), R5
	MOVW	a4+16(FP), R6
	MOVW	a5+20(FP), R7
	MOVW	a6+24(FP), R8
	SYSCALL	R31
	BVC	ok
	MOVW	$-1, R5
	MOVW	R5, r1+28(FP)
	MOVW	R0, r2+32(FP)
	MOVW	R3, errno+36(FP)
	RET
ok:
	MOVW	R3, r1+28(FP)
	MOVW	R4, r2+32(FP)
	MOVW	R0, errno+36(FP)
	RET
