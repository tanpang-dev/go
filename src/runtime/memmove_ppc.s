// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// See memmove Go doc for important implementation constraints.

// func memmove(to, from unsafe.Pointer, n uintptr)
//
// Forward copies with equally 8-byte-aligned pointers move 32 bytes per
// iteration through FPRs (lfd/stfd move the bits untouched). Go code on ppc
// uses soft float, so F0-F3 are free here. Equally 4-byte-aligned copies use
// an unrolled word loop. Everything else, and backward copies, is word or
// byte at a time.
TEXT runtime·memmove(SB),NOSPLIT|NOFRAME,$0-12
	MOVD	to+0(FP), R3
	MOVD	from+4(FP), R4
	MOVWZ	n+8(FP), R5
	CMP	R5, $0
	BEQ	done
	CMPU	R3, R4
	BEQ	done
	BLT	forward
	ADD	R4, R5, R6	// from+n
	CMPU	R3, R6
	BLT	backward	// overlapping with to > from

forward:
	CMPU	R5, $32
	BLT	fwords
	XOR	R3, R4, R6
	ANDCC	$7, R6, R7
	BEQ	f8
	ANDCC	$3, R6, R7
	BEQ	f4
	BR	fbytes

f8:	// align both to 8
	ANDCC	$7, R3, R7
	BEQ	f8loop
	MOVBZ	0(R4), R6
	MOVB	R6, 0(R3)
	ADD	$1, R3
	ADD	$1, R4
	SUB	$1, R5
	BR	f8
f8loop:
	CMPU	R5, $32
	BLT	fwords
	FMOVD	0(R4), F0
	FMOVD	8(R4), F1
	FMOVD	16(R4), F2
	FMOVD	24(R4), F3
	FMOVD	F0, 0(R3)
	FMOVD	F1, 8(R3)
	FMOVD	F2, 16(R3)
	FMOVD	F3, 24(R3)
	ADD	$32, R3
	ADD	$32, R4
	SUB	$32, R5
	BR	f8loop

f4:	// align both to 4
	ANDCC	$3, R3, R7
	BEQ	f4loop
	MOVBZ	0(R4), R6
	MOVB	R6, 0(R3)
	ADD	$1, R3
	ADD	$1, R4
	SUB	$1, R5
	BR	f4
f4loop:
	CMPU	R5, $32
	BLT	fwords
	MOVWZ	0(R4), R6
	MOVWZ	4(R4), R7
	MOVWZ	8(R4), R8
	MOVWZ	12(R4), R9
	MOVW	R6, 0(R3)
	MOVW	R7, 4(R3)
	MOVW	R8, 8(R3)
	MOVW	R9, 12(R3)
	MOVWZ	16(R4), R6
	MOVWZ	20(R4), R7
	MOVWZ	24(R4), R8
	MOVWZ	28(R4), R9
	MOVW	R6, 16(R3)
	MOVW	R7, 20(R3)
	MOVW	R8, 24(R3)
	MOVW	R9, 28(R3)
	ADD	$32, R3
	ADD	$32, R4
	SUB	$32, R5
	BR	f4loop

fwords:	// short or tail: words only if both aligned
	OR	R3, R4, R6
	ANDCC	$3, R6, R6
	BNE	fbytes
fw:
	CMPU	R5, $4
	BLT	fbytes
	MOVWZ	0(R4), R6
	MOVW	R6, 0(R3)
	ADD	$4, R3
	ADD	$4, R4
	SUB	$4, R5
	BR	fw
fbytes:
	CMP	R5, $0
	BEQ	done
	MOVBZ	0(R4), R6
	MOVB	R6, 0(R3)
	ADD	$1, R3
	ADD	$1, R4
	SUB	$1, R5
	BR	fbytes

backward:
	// to > from: copy from the end
	ADD	R5, R3
	ADD	R5, R4
	OR	R3, R4, R6
	ANDCC	$3, R6, R6
	BNE	bbytes
bwords:
	CMPU	R5, $4
	BLT	bbytes
	SUB	$4, R3
	SUB	$4, R4
	MOVWZ	0(R4), R6
	MOVW	R6, 0(R3)
	SUB	$4, R5
	BR	bwords
bbytes:
	CMP	R5, $0
	BEQ	done
	SUB	$1, R3
	SUB	$1, R4
	MOVBZ	0(R4), R6
	MOVB	R6, 0(R3)
	SUB	$1, R5
	BR	bbytes
done:
	RET
