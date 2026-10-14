// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// func IndexByte(b []byte, c byte) int
TEXT ·IndexByte(SB),NOSPLIT,$0-20
	MOVW	b_base+0(FP), R3
	MOVW	b_len+4(FP), R4
	MOVBZ	c+12(FP), R5
	BL	indexbytebody<>(SB)
	MOVW	R3, ret+16(FP)
	RET

// func IndexByteString(s string, c byte) int
TEXT ·IndexByteString(SB),NOSPLIT,$0-16
	MOVW	s_base+0(FP), R3
	MOVW	s_len+4(FP), R4
	MOVBZ	c+8(FP), R5
	BL	indexbytebody<>(SB)
	MOVW	R3, ret+12(FP)
	RET

// indexbytebody finds byte R5 in [R3, R3+R4) and returns its index or -1
// in R3. Aligned words are tested 4 bytes at a time: after XOR with the
// replicated byte, ^(((x & 0x7f7f7f7f) + 0x7f7f7f7f) | x | 0x7f7f7f7f) has
// 0x80 in exactly the zero (matching) bytes, and on big-endian the first
// match is the count of leading zeros / 8.
TEXT indexbytebody<>(SB),NOSPLIT|NOFRAME,$0
	MOVD	R3, R10	// start
	ADD	R3, R4, R9	// end
	MOVW	$0x01010101, R6
	MULLW	R5, R6, R11	// replicated byte
	MOVW	$0x7f7f7f7f, R7
head:
	CMPU	R3, R9
	BGE	notfound
	ANDCC	$3, R3, R6
	BEQ	words
	MOVBZ	0(R3), R6
	CMPW	R6, R5
	BEQ	found
	ADD	$1, R3
	BR	head
words:
	SUB	R3, R9, R6
	CMPU	R6, $4
	BLT	tail
	MOVWZ	0(R3), R6
	XOR	R11, R6
	AND	R7, R6, R8
	ADD	R7, R8
	OR	R6, R8
	OR	R7, R8
	NOR	R8, R8, R8
	CMPW	R8, $0
	BNE	foundword
	ADD	$4, R3
	BR	words
foundword:
	CNTLZW	R8, R8
	SRW	$3, R8
	ADD	R8, R3
	BR	found
tail:
	CMPU	R3, R9
	BGE	notfound
	MOVBZ	0(R3), R6
	CMPW	R6, R5
	BEQ	found
	ADD	$1, R3
	BR	tail
found:
	SUB	R10, R3
	RET
notfound:
	MOVW	$-1, R3
	RET
