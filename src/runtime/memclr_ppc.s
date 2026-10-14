// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// See memclrNoHeapPointers Go doc for important implementation constraints.

// func memclrNoHeapPointers(ptr unsafe.Pointer, n uintptr)
// Bytes up to word alignment, words up to the 32-byte cache line, then dcbz
// clears whole lines (the 750CL's L1 line is 32 bytes), then words and bytes.
TEXT runtime·memclrNoHeapPointers(SB),NOSPLIT|NOFRAME,$0-8
	MOVD	ptr+0(FP), R3
	MOVWZ	n+4(FP), R4
	ADD	R3, R4, R5	// R5 = end
	CMPU	R4, $64
	BLT	words	// short: no dcbz
head:
	ANDCC	$3, R3, R6
	BEQ	walign
	MOVB	R0, 0(R3)
	ADD	$1, R3
	BR	head
walign:
	ANDCC	$31, R3, R6
	BEQ	lines
	MOVW	R0, 0(R3)
	ADD	$4, R3
	BR	walign
lines:
	RLWNM	$0, R5, $~31, R6	// R6 = end of the last whole line
dcbz:
	CMPU	R3, R6
	BGE	words
	DCBZ	(R3)
	ADD	$32, R3
	BR	dcbz
words:
	// Short clears may start unaligned; those go byte by byte.
	ANDCC	$3, R3, R6
	BNE	tail
	SUB	R3, R5, R4
	CMPU	R4, $4
	BLT	tail
	MOVW	R0, 0(R3)
	ADD	$4, R3
	BR	words
tail:
	CMPU	R3, R5
	BGE	done
	MOVB	R0, 0(R3)
	ADD	$1, R3
	BR	tail
done:
	RET
