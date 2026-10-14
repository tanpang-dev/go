// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

.file "gcc_ppc.S"

/*
 * void crosscall1(void (*fn)(void), void (*setg_gcc)(void *g), void *g)
 *
 * Calling into the gc tool chain, where all registers are caller save.
 * Called from the 32-bit SysV ABI, where r14-r31, f14-f31 and cr2-cr4
 * are callee-save, so they must be saved explicitly, along with LR.
 * Frame: 0 back chain, 8-79 r14-r31, 80 cr, 96-239 f14-f31.
 */
.globl crosscall1
.type crosscall1, @function
crosscall1:
	stwu	1, -240(1)
	mflr	0
	stw	0, 244(1)
	stmw	14, 8(1)
	mfcr	0
	stw	0, 80(1)
	stfd	14, 96(1)
	stfd	15, 104(1)
	stfd	16, 112(1)
	stfd	17, 120(1)
	stfd	18, 128(1)
	stfd	19, 136(1)
	stfd	20, 144(1)
	stfd	21, 152(1)
	stfd	22, 160(1)
	stfd	23, 168(1)
	stfd	24, 176(1)
	stfd	25, 184(1)
	stfd	26, 192(1)
	stfd	27, 200(1)
	stfd	28, 208(1)
	stfd	29, 216(1)
	stfd	30, 224(1)
	stfd	31, 232(1)

	mr	14, 3		// fn
	mr	15, 5		// g
	mtctr	4
	mr	3, 5
	bctrl			// setg_gcc(g)
	li	0, 0		// Go keeps r0 zero
	mr	30, 15		// and g in r30
	mtctr	14
	bctrl			// fn()

	lfd	14, 96(1)
	lfd	15, 104(1)
	lfd	16, 112(1)
	lfd	17, 120(1)
	lfd	18, 128(1)
	lfd	19, 136(1)
	lfd	20, 144(1)
	lfd	21, 152(1)
	lfd	22, 160(1)
	lfd	23, 168(1)
	lfd	24, 176(1)
	lfd	25, 184(1)
	lfd	26, 192(1)
	lfd	27, 200(1)
	lfd	28, 208(1)
	lfd	29, 216(1)
	lfd	30, 224(1)
	lfd	31, 232(1)
	lwz	0, 80(1)
	mtcrf	0xff, 0
	lmw	14, 8(1)
	lwz	0, 244(1)
	mtlr	0
	addi	1, 1, 240
	blr
.size crosscall1, .-crosscall1

#ifdef __ELF__
.section .note.GNU-stack,"",%progbits
#endif
