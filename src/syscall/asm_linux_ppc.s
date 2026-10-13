// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && ppc

#include "textflag.h"
#include "funcdata.h"

//
// System calls for 32-bit PowerPC, Linux
//

// func Syscall9(trap, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2, err uintptr)
// The kernel takes at most six register arguments; a7-a9 are ignored.
TEXT ·Syscall9(SB),NOSPLIT,$0-52
	NO_LOCAL_POINTERS
	BL	runtime·entersyscall(SB)
	MOVW	a1+4(FP), R3
	MOVW	a2+8(FP), R4
	MOVW	a3+12(FP), R5
	MOVW	a4+16(FP), R6
	MOVW	a5+20(FP), R7
	MOVW	a6+24(FP), R8
	MOVW	trap+0(FP), R9	// syscall entry
	SYSCALL	R9
	BVC	ok9
	MOVW	$-1, R4
	MOVW	R4, r1+40(FP)	// r1
	MOVW	R0, r2+44(FP)	// r2
	MOVW	R3, err+48(FP)	// errno
	BL	runtime·exitsyscall(SB)
	RET
ok9:
	MOVW	R3, r1+40(FP)	// r1
	MOVW	R4, r2+44(FP)	// r2
	MOVW	R0, err+48(FP)	// errno
	BL	runtime·exitsyscall(SB)
	RET

// func rawVforkSyscall(trap, a1, a2, a3 uintptr) (r1, err uintptr)
TEXT ·rawVforkSyscall(SB),NOSPLIT|NOFRAME,$0-24
	MOVW	a1+4(FP), R3
	MOVW	a2+8(FP), R4
	MOVW	a3+12(FP), R5
	MOVW	R0, R6
	MOVW	R0, R7
	MOVW	R0, R8
	MOVW	trap+0(FP), R9	// syscall entry
	SYSCALL	R9
	BVC	ok
	MOVW	$-1, R4
	MOVW	R4, r1+16(FP)	// r1
	MOVW	R3, err+20(FP)	// errno
	RET
ok:
	MOVW	R3, r1+16(FP)	// r1
	MOVW	R0, err+20(FP)	// errno
	RET

// func rawSyscallNoError(trap, a1, a2, a3 uintptr) (r1, r2 uintptr)
TEXT ·rawSyscallNoError(SB),NOSPLIT|NOFRAME,$0-24
	MOVW	a1+4(FP), R3
	MOVW	a2+8(FP), R4
	MOVW	a3+12(FP), R5
	MOVW	R0, R6
	MOVW	R0, R7
	MOVW	R0, R8
	MOVW	trap+0(FP), R9	// syscall entry
	SYSCALL	R9
	MOVW	R3, r1+16(FP)	// r1
	MOVW	R4, r2+20(FP)	// r2
	RET
