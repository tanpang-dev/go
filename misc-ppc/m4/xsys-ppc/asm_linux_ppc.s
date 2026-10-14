// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && ppc && gc

#include "textflag.h"

//
// System calls for 32-bit PowerPC, Linux
//

// Just jump to package syscall's implementation for all these functions.
// The runtime may know about them.

TEXT ·Syscall(SB),NOSPLIT,$0-28
	JMP	syscall·Syscall(SB)

TEXT ·Syscall6(SB),NOSPLIT,$0-40
	JMP	syscall·Syscall6(SB)

TEXT ·Syscall9(SB),NOSPLIT,$0-52
	JMP	syscall·Syscall9(SB)

TEXT ·SyscallNoError(SB),NOSPLIT,$0-24
	BL	runtime·entersyscall(SB)
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
	BL	runtime·exitsyscall(SB)
	RET

TEXT ·RawSyscall(SB),NOSPLIT,$0-28
	JMP	syscall·RawSyscall(SB)

TEXT ·RawSyscall6(SB),NOSPLIT,$0-40
	JMP	syscall·RawSyscall6(SB)

TEXT ·RawSyscallNoError(SB),NOSPLIT,$0-24
	MOVW	a1+4(FP), R3
	MOVW	a2+8(FP), R4
	MOVW	a3+12(FP), R5
	MOVW	R0, R6
	MOVW	R0, R7
	MOVW	R0, R8
	MOVW	trap+0(FP), R9	// syscall entry
	SYSCALL	R9
	MOVW	R3, r1+16(FP)
	MOVW	R4, r2+20(FP)
	RET
