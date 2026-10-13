// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"

// 32-bit PowerPC (750CL). There is no lwsync and no byte-sized lbarx/stbcx.,
// so barriers are full SYNCs and 8-bit operations work on the containing
// big-endian word. 64-bit operations use the spinlock in atomic_ppc.go.

// func Cas(ptr *uint32, old, new uint32) bool
TEXT ·Cas(SB),NOSPLIT,$0-13
	MOVW	ptr+0(FP), R3
	MOVW	old+4(FP), R4
	MOVW	new+8(FP), R5
	SYNC
cas_again:
	LWAR	(R3), R6
	CMPW	R6, R4
	BNE	cas_fail
	STWCCC	R5, (R3)
	BNE	cas_again
	MOVW	$1, R3
	SYNC
	MOVB	R3, ret+12(FP)
	RET
cas_fail:
	SYNC
	MOVB	R0, ret+12(FP)
	RET

TEXT ·Store(SB),NOSPLIT,$0-8
	MOVW	ptr+0(FP), R3
	MOVW	val+4(FP), R4
	SYNC
	MOVW	R4, 0(R3)
	SYNC
	RET

TEXT ·Store8(SB),NOSPLIT,$0-5
	MOVW	ptr+0(FP), R3
	MOVB	val+4(FP), R4
	SYNC
	MOVB	R4, 0(R3)
	SYNC
	RET

TEXT ·Load(SB),NOSPLIT,$0-8
	MOVW	ptr+0(FP), R3
	SYNC
	MOVWZ	0(R3), R3
	SYNC
	MOVW	R3, ret+4(FP)
	RET

TEXT ·Load8(SB),NOSPLIT,$0-5
	MOVW	ptr+0(FP), R3
	SYNC
	MOVBZ	0(R3), R3
	SYNC
	MOVB	R3, ret+4(FP)
	RET

// uint32 Xadd(uint32 volatile *val, int32 delta)
// Atomically:
//	*val += delta;
//	return *val;
TEXT ·Xadd(SB),NOSPLIT,$0-12
	MOVW	ptr+0(FP), R4
	MOVW	delta+4(FP), R5
	SYNC
	LWAR	(R4), R3
	ADD	R5, R3
	STWCCC	R3, (R4)
	BNE	-3(PC)
	SYNC
	MOVW	R3, ret+8(FP)
	RET

// uint32 Xchg(ptr *uint32, new uint32)
// Atomically:
//	old := *ptr;
//	*ptr = new;
//	return old;
TEXT ·Xchg(SB),NOSPLIT,$0-12
	MOVW	ptr+0(FP), R4
	MOVW	new+4(FP), R5
	SYNC
	LWAR	(R4), R3
	STWCCC	R5, (R4)
	BNE	-2(PC)
	SYNC
	MOVW	R3, ret+8(FP)
	RET

// byteLane sets R6 = ptr&^3 and R7 = the left shift that moves a byte to its
// lane in that big-endian word: (3 - ptr&3) * 8.
#define byteLane \
	RLWNM	$0, R3, $~3, R6; \
	ANDCC	$3, R3, R7; \
	XOR	$3, R7; \
	SLW	$3, R7

// uint8 Xchg8(ptr *uint8, new uint8)
TEXT ·Xchg8(SB),NOSPLIT,$0-9
	MOVW	ptr+0(FP), R3
	MOVBZ	new+4(FP), R4
	byteLane
	MOVW	$0xff, R8
	SLW	R7, R8	// R8 = lane mask
	SLW	R7, R4	// R4 = new byte in its lane
	SYNC
again:
	LWAR	(R6), R9
	ANDN	R8, R9, R10	// clear the lane
	OR	R4, R10
	STWCCC	R10, (R6)
	BNE	again
	SYNC
	SRW	R7, R9
	MOVB	R9, ret+8(FP)
	RET

// void Or8(byte volatile*, byte);
TEXT ·Or8(SB),NOSPLIT,$0-5
	MOVW	ptr+0(FP), R3
	MOVBZ	val+4(FP), R4
	byteLane
	SLW	R7, R4
	SYNC
again:
	LWAR	(R6), R9
	OR	R4, R9
	STWCCC	R9, (R6)
	BNE	again
	SYNC
	RET

// void And8(byte volatile*, byte);
TEXT ·And8(SB),NOSPLIT,$0-5
	MOVW	ptr+0(FP), R3
	MOVBZ	val+4(FP), R4
	byteLane
	MOVW	$0xff, R8
	SLW	R7, R8
	SLW	R7, R4
	ORN	R8, R4	// other lanes all ones
	SYNC
again:
	LWAR	(R6), R9
	AND	R4, R9
	STWCCC	R9, (R6)
	BNE	again
	SYNC
	RET

// func Or(addr *uint32, v uint32)
TEXT ·Or(SB),NOSPLIT,$0-8
	MOVW	ptr+0(FP), R3
	MOVW	val+4(FP), R4
	SYNC
again:
	LWAR	(R3), R6
	OR	R4, R6
	STWCCC	R6, (R3)
	BNE	again
	SYNC
	RET

// func And(addr *uint32, v uint32)
TEXT ·And(SB),NOSPLIT,$0-8
	MOVW	ptr+0(FP), R3
	MOVW	val+4(FP), R4
	SYNC
again:
	LWAR	(R3), R6
	AND	R4, R6
	STWCCC	R6, (R3)
	BNE	again
	SYNC
	RET

// func Or32(addr *uint32, v uint32) old uint32
TEXT ·Or32(SB),NOSPLIT,$0-12
	MOVW	ptr+0(FP), R3
	MOVW	val+4(FP), R4
	SYNC
again:
	LWAR	(R3), R6
	OR	R4, R6, R7
	STWCCC	R7, (R3)
	BNE	again
	SYNC
	MOVW	R6, ret+8(FP)
	RET

// func And32(addr *uint32, v uint32) old uint32
TEXT ·And32(SB),NOSPLIT,$0-12
	MOVW	ptr+0(FP), R3
	MOVW	val+4(FP), R4
	SYNC
again:
	LWAR	(R3), R6
	AND	R4, R6, R7
	STWCCC	R7, (R3)
	BNE	again
	SYNC
	MOVW	R6, ret+8(FP)
	RET

// func spinLock(state *uint32)
TEXT ·spinLock(SB),NOSPLIT,$0-4
	MOVW	state+0(FP), R3
	MOVW	$1, R4
again:
	LWAR	(R3), R5
	CMPW	R5, $0
	BNE	again
	STWCCC	R4, (R3)
	BNE	again
	SYNC
	RET

// func spinUnlock(state *uint32)
TEXT ·spinUnlock(SB),NOSPLIT,$0-4
	MOVW	state+0(FP), R3
	SYNC
	MOVW	R0, 0(R3)
	RET

TEXT ·Casint32(SB),NOSPLIT,$0-13
	JMP	·Cas(SB)

TEXT ·Casint64(SB),NOSPLIT,$0-21
	JMP	·Cas64(SB)

TEXT ·Casuintptr(SB),NOSPLIT,$0-13
	JMP	·Cas(SB)

TEXT ·CasRel(SB),NOSPLIT,$0-13
	JMP	·Cas(SB)

TEXT ·Loaduintptr(SB),NOSPLIT,$0-8
	JMP	·Load(SB)

TEXT ·Loaduint(SB),NOSPLIT,$0-8
	JMP	·Load(SB)

TEXT ·Loadp(SB),NOSPLIT,$-0-8
	JMP	·Load(SB)

TEXT ·Storeint32(SB),NOSPLIT,$0-8
	JMP	·Store(SB)

TEXT ·Storeint64(SB),NOSPLIT,$0-12
	JMP	·Store64(SB)

TEXT ·Storeuintptr(SB),NOSPLIT,$0-8
	JMP	·Store(SB)

TEXT ·Xadduintptr(SB),NOSPLIT,$0-12
	JMP	·Xadd(SB)

TEXT ·Loadint32(SB),NOSPLIT,$0-8
	JMP	·Load(SB)

TEXT ·Loadint64(SB),NOSPLIT,$0-12
	JMP	·Load64(SB)

TEXT ·Xaddint32(SB),NOSPLIT,$0-12
	JMP	·Xadd(SB)

TEXT ·Xaddint64(SB),NOSPLIT,$0-20
	JMP	·Xadd64(SB)

TEXT ·Casp1(SB),NOSPLIT,$0-13
	JMP	·Cas(SB)

TEXT ·Xchgint32(SB),NOSPLIT,$0-12
	JMP	·Xchg(SB)

TEXT ·Xchgint64(SB),NOSPLIT,$0-20
	JMP	·Xchg64(SB)

TEXT ·Xchguintptr(SB),NOSPLIT,$0-12
	JMP	·Xchg(SB)

TEXT ·StorepNoWB(SB),NOSPLIT,$0-8
	JMP	·Store(SB)

TEXT ·StoreRel(SB),NOSPLIT,$0-8
	JMP	·Store(SB)

TEXT ·StoreReluintptr(SB),NOSPLIT,$0-8
	JMP	·Store(SB)

TEXT ·LoadAcq(SB),NOSPLIT,$0-8
	JMP	·Load(SB)

TEXT ·LoadAcquintptr(SB),NOSPLIT,$0-8
	JMP	·Load(SB)

TEXT ·Anduintptr(SB),NOSPLIT,$0-12
	JMP	·And32(SB)

TEXT ·Oruintptr(SB),NOSPLIT,$0-12
	JMP	·Or32(SB)
