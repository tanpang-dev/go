// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// memeqbody compares R3 and R4 for R5 bytes and leaves 1 or 0 in R3.
// Equally word-aligned inputs compare 16 bytes per iteration.
TEXT memeqbody<>(SB),NOSPLIT|NOFRAME,$0
	CMPU	R3, R4
	BEQ	eq
	ADD	R3, R5, R6	// end of a
	OR	R3, R4, R7
	ANDCC	$3, R7, R7
	BNE	bytes
words16:
	SUB	R3, R6, R7
	CMPU	R7, $16
	BLT	words4
	MOVWZ	0(R3), R7
	MOVWZ	0(R4), R8
	CMPW	R7, R8
	BNE	ne
	MOVWZ	4(R3), R7
	MOVWZ	4(R4), R8
	CMPW	R7, R8
	BNE	ne
	MOVWZ	8(R3), R7
	MOVWZ	8(R4), R8
	CMPW	R7, R8
	BNE	ne
	MOVWZ	12(R3), R7
	MOVWZ	12(R4), R8
	CMPW	R7, R8
	BNE	ne
	ADD	$16, R3
	ADD	$16, R4
	BR	words16
words4:
	SUB	R3, R6, R7
	CMPU	R7, $4
	BLT	bytes
	MOVWZ	0(R3), R7
	MOVWZ	0(R4), R8
	CMPW	R7, R8
	BNE	ne
	ADD	$4, R3
	ADD	$4, R4
	BR	words4
bytes:
	CMPU	R3, R6
	BEQ	eq
	MOVBZ	0(R3), R7
	MOVBZ	0(R4), R8
	ADD	$1, R3
	ADD	$1, R4
	CMPW	R7, R8
	BEQ	bytes
ne:
	MOVW	$0, R3
	RET
eq:
	MOVW	$1, R3
	RET

// func memequal(a, b unsafe.Pointer, size uintptr) bool
TEXT runtime·memequal(SB),NOSPLIT,$0-13
	MOVW	a+0(FP), R3
	MOVW	b+4(FP), R4
	MOVW	size+8(FP), R5
	BL	memeqbody<>(SB)
	MOVB	R3, ret+12(FP)
	RET

// memequal_varlen(a, b unsafe.Pointer) bool
TEXT runtime·memequal_varlen(SB),NOSPLIT,$0-9
	MOVW	a+0(FP), R3
	MOVW	b+4(FP), R4
	MOVW	4(R11), R5	// compiler stores size at offset 4 in the closure
	BL	memeqbody<>(SB)
	MOVB	R3, ret+8(FP)
	RET
