// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// See memmove Go doc for important implementation constraints.

// func memmove(to, from unsafe.Pointer, n uintptr)
// ponytail: word loop when both pointers are word aligned, bytes otherwise.
TEXT runtime·memmove(SB),NOSPLIT|NOFRAME,$0-12
	MOVD	to+0(FP), R3
	MOVD	from+4(FP), R4
	MOVWZ	n+8(FP), R5
	CMP	R5, $0
	BEQ	done
	CMPU	R3, R4
	BEQ	done
	BGT	backward

	// forward copy (to < from)
	OR	R3, R4, R6
	ANDCC	$3, R6, R6
	BNE	fbytes
fwords:
	CMPU	R5, $4
	BLT	fbytes
	MOVWZ	0(R4), R6
	MOVW	R6, 0(R3)
	ADD	$4, R3
	ADD	$4, R4
	SUB	$4, R5
	BR	fwords
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
