// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// See memclrNoHeapPointers Go doc for important implementation constraints.

// func memclrNoHeapPointers(ptr unsafe.Pointer, n uintptr)
// ponytail: byte loop up to word alignment, then words; a dcbz-based loop is
// the upgrade if clearing shows up in profiles.
TEXT runtime·memclrNoHeapPointers(SB),NOSPLIT|NOFRAME,$0-8
	MOVD	ptr+0(FP), R3
	MOVWZ	n+4(FP), R4
	ADD	R3, R4, R5	// R5 = end
head:
	CMP	R4, $0
	BEQ	done
	ANDCC	$3, R3, R6
	BEQ	words
	MOVB	R0, 0(R3)
	ADD	$1, R3
	SUB	$1, R4
	BR	head
words:
	CMPU	R4, $4
	BLT	tail
	MOVW	R0, 0(R3)
	ADD	$4, R3
	SUB	$4, R4
	BR	words
tail:
	CMPU	R3, R5
	BGE	done
	MOVB	R0, 0(R3)
	ADD	$1, R3
	BR	tail
done:
	RET
