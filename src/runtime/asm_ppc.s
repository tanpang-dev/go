// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// 32-bit PowerPC. Structure follows asm_mipsx.s (stack-based ABI0), idioms
// follow asm_ppc64x.s. MOVD loads/stores are encoded as word operations by
// the assembler for GOARCH=ppc.
//
// Registers: R0 is kept zero, R1 is SP, R2 is the thread pointer (never
// touched), R11 is the closure context, R30 is g, R31 is REGTMP.
// The fixed frame is 8 bytes: Go's saved LR at 0(R1), SysV LR slot at 4(R1).

#include "go_asm.h"
#include "go_tls.h"
#include "funcdata.h"
#include "textflag.h"

#define FIXED_FRAME 8

TEXT runtime·rt0_go(SB),NOSPLIT|TOPFRAME,$0
	// R1 = stack; R3 = argc; R4 = argv
	MOVW	$0, R0

	SUB	$(FIXED_FRAME+8), R1
	MOVW	R3, FIXED_FRAME+0(R1)	// argc
	MOVW	R4, FIXED_FRAME+4(R1)	// argv

	// create istack out of the given (operating system) stack.
	MOVD	$runtime·g0(SB), g
	MOVD	$(-64*1024), R31
	ADD	R31, R1, R3
	MOVD	R3, g_stackguard0(g)
	MOVD	R3, g_stackguard1(g)
	MOVD	R3, (g_stack+stack_lo)(g)
	MOVD	R1, (g_stack+stack_hi)(g)

	// cgo is not supported on linux/ppc yet.
	MOVD	_cgo_init(SB), R12
	CMP	R12, $0
	BEQ	nocgo
	BL	runtime·abort(SB)

nocgo:
	// update stackguard after _cgo_init
	MOVD	(g_stack+stack_lo)(g), R3
	ADD	$const_stackGuard, R3
	MOVD	R3, g_stackguard0(g)
	MOVD	R3, g_stackguard1(g)

	// set the per-goroutine and per-mach "registers"
	MOVD	$runtime·m0(SB), R3

	// save m->g0 = g0
	MOVD	g, m_g0(R3)
	// save m0 to g0->m
	MOVD	R3, g_m(g)

	BL	runtime·check(SB)

	// args are already prepared
	BL	runtime·args(SB)
	BL	runtime·osinit(SB)
	BL	runtime·schedinit(SB)

	// create a new goroutine to start program
	MOVD	$runtime·mainPC(SB), R3	// entry
	MOVW	R3, FIXED_FRAME+0(R1)
	BL	runtime·newproc(SB)

	// start this M
	BL	runtime·mstart(SB)

	UNDEF
	RET

DATA	runtime·mainPC+0(SB)/4,$runtime·main(SB)
GLOBL	runtime·mainPC(SB),RODATA,$4

TEXT runtime·breakpoint(SB),NOSPLIT|NOFRAME,$0-0
	TW	$31, R0, R0
	RET

TEXT runtime·asminit(SB),NOSPLIT|NOFRAME,$0-0
	RET

TEXT runtime·mstart(SB),NOSPLIT|TOPFRAME,$0
	BL	runtime·mstart0(SB)
	RET // not reached

/*
 *  go-routine
 */

// void gogo(Gobuf*)
// restore state from Gobuf; longjmp
TEXT runtime·gogo(SB),NOSPLIT|NOFRAME,$0-4
	MOVD	buf+0(FP), R5
	MOVD	gobuf_g(R5), R6
	MOVD	0(R6), R4	// make sure g != nil
	BR	gogo<>(SB)

TEXT gogo<>(SB),NOSPLIT|NOFRAME,$0
	MOVD	R6, g
	BL	runtime·save_g(SB)

	MOVD	gobuf_sp(R5), R1
	MOVD	gobuf_lr(R5), R31
	MOVD	R31, LR
	MOVD	gobuf_ctxt(R5), R11
	MOVD	R0, gobuf_sp(R5)
	MOVD	R0, gobuf_lr(R5)
	MOVD	R0, gobuf_ctxt(R5)
	CMP	R0, R0 // set condition codes for == test, needed by stack split
	MOVD	gobuf_pc(R5), R12
	MOVD	R12, CTR
	BR	(CTR)

// void mcall(fn func(*g))
// Switch to m->g0's stack, call fn(g).
// Fn must never return. It should gogo(&g->sched)
// to keep running g.
TEXT runtime·mcall(SB),NOSPLIT|NOFRAME,$0-4
	MOVD	fn+0(FP), R11	// context
	// Save caller state in g->sched
	MOVD	R1, (g_sched+gobuf_sp)(g)
	MOVD	LR, R31
	MOVD	R31, (g_sched+gobuf_pc)(g)
	MOVD	R0, (g_sched+gobuf_lr)(g)

	// Switch to m->g0 & its stack, call fn.
	MOVD	g, R3
	MOVD	g_m(g), R8
	MOVD	m_g0(R8), g
	BL	runtime·save_g(SB)
	CMP	g, R3
	BNE	2(PC)
	BR	runtime·badmcall(SB)
	MOVD	0(R11), R12	// code pointer
	MOVD	R12, CTR
	MOVD	(g_sched+gobuf_sp)(g), R1	// sp = m->g0->sched.sp
	SUB	$(FIXED_FRAME+8), R1	// make room for 1 arg and a fake LR
	MOVW	R3, FIXED_FRAME+0(R1)
	MOVW	R0, 0(R1)
	BL	(CTR)
	BR	runtime·badmcall2(SB)

// systemstack_switch is a dummy routine that systemstack leaves at the bottom
// of the G stack. We need to distinguish the routine that
// lives at the bottom of the G stack from the one that lives
// at the top of the system stack because the one at the top of
// the system stack terminates the stack walk (see topofstack()).
TEXT runtime·systemstack_switch(SB),NOSPLIT,$0-0
	UNDEF
	UNDEF
	BL	(LR)	// make sure this function is not leaf
	RET

// func systemstack(fn func())
TEXT runtime·systemstack(SB),NOSPLIT,$0-4
	MOVD	fn+0(FP), R3	// R3 = fn
	MOVD	R3, R11	// context
	MOVD	g_m(g), R4	// R4 = m

	MOVD	m_gsignal(R4), R5	// R5 = gsignal
	CMP	g, R5
	BEQ	noswitch

	MOVD	m_g0(R4), R5	// R5 = g0
	CMP	g, R5
	BEQ	noswitch

	MOVD	m_curg(R4), R6
	CMP	g, R6
	BEQ	switch

	// Bad: g is not gsignal, not g0, not curg. What is it?
	// Hide call from linker nosplit analysis.
	MOVD	$runtime·badsystemstack(SB), R12
	MOVD	R12, CTR
	BL	(CTR)
	BL	runtime·abort(SB)

switch:
	// save our state in g->sched. Pretend to
	// be systemstack_switch if the G stack is scanned.
	BL	gosave_systemstack_switch<>(SB)

	// switch to g0
	MOVD	R5, g
	BL	runtime·save_g(SB)
	MOVD	(g_sched+gobuf_sp)(g), R1

	// call target function
	MOVD	0(R11), R12	// code pointer
	MOVD	R12, CTR
	BL	(CTR)

	// switch back to g
	MOVD	g_m(g), R3
	MOVD	m_curg(R3), g
	BL	runtime·save_g(SB)
	MOVD	(g_sched+gobuf_sp)(g), R1
	MOVD	R0, (g_sched+gobuf_sp)(g)
	RET

noswitch:
	// already on m stack, just call directly
	MOVD	0(R11), R12	// code pointer
	MOVD	R12, CTR
	BL	(CTR)
	RET

// func switchToCrashStack0(fn func())
TEXT runtime·switchToCrashStack0(SB),NOSPLIT,$0-4
	MOVD	fn+0(FP), R11	// context register
	MOVD	g_m(g), R3	// curm

	// set g to gcrash
	MOVD	$runtime·gcrash(SB), g	// g = &gcrash
	BL	runtime·save_g(SB)
	MOVD	R3, g_m(g)	// g.m = curm
	MOVD	g, m_g0(R3)	// curm.g0 = g

	// switch to crashstack
	MOVD	(g_stack+stack_hi)(g), R3
	ADD	$(-4*8), R3, R1

	// call target function
	MOVD	0(R11), R12
	MOVD	R12, CTR
	BL	(CTR)

	// should never return
	BL	runtime·abort(SB)
	UNDEF

/*
 * support for morestack
 */

// Called during function prolog when more stack is needed.
// Caller has already loaded:
// R3: framesize, R4: argsize, R5: LR
//
// The traceback routines see morestack on a g0 as being
// the top of a stack (for example, morestack calling newstack
// calling the scheduler calling newm calling gc), so we must
// record an argument size. For that purpose, it has no arguments.
TEXT runtime·morestack(SB),NOSPLIT|NOFRAME,$0-0
	// Called from f.
	// Set g->sched to context in f.
	MOVD	R1, (g_sched+gobuf_sp)(g)
	MOVD	LR, R8
	MOVD	R8, (g_sched+gobuf_pc)(g)
	MOVD	R5, (g_sched+gobuf_lr)(g)
	MOVD	R11, (g_sched+gobuf_ctxt)(g)

	// Cannot grow scheduler stack (m->g0).
	MOVD	g_m(g), R7
	MOVD	m_g0(R7), R8
	CMP	g, R8
	BNE	3(PC)
	BL	runtime·badmorestackg0(SB)
	BL	runtime·abort(SB)

	// Cannot grow signal stack (m->gsignal).
	MOVD	m_gsignal(R7), R8
	CMP	g, R8
	BNE	3(PC)
	BL	runtime·badmorestackgsignal(SB)
	BL	runtime·abort(SB)

	// Called from f.
	// Set m->morebuf to f's caller.
	MOVD	R5, (m_morebuf+gobuf_pc)(R7)	// f's caller's PC
	MOVD	R1, (m_morebuf+gobuf_sp)(R7)	// f's caller's SP
	MOVD	g, (m_morebuf+gobuf_g)(R7)

	// Call newstack on m->g0's stack.
	MOVD	m_g0(R7), g
	BL	runtime·save_g(SB)
	MOVD	(g_sched+gobuf_sp)(g), R1
	MOVDU	R0, -(FIXED_FRAME+0)(R1)	// create a call frame on g0
	BL	runtime·newstack(SB)

	// Not reached, but make sure the return PC from the call to newstack
	// is still in this function, and not the beginning of the next.
	UNDEF

TEXT runtime·morestack_noctxt(SB),NOSPLIT|NOFRAME,$0-0
	// Force SPWRITE. This function doesn't actually write SP,
	// but it is called with a special calling convention where
	// the caller doesn't save LR on stack but passes it as a
	// register (R5), and the unwinder currently doesn't understand.
	// Make it SPWRITE to stop unwinding. (See issue 54332)
	OR	R0, R1
	MOVD	R0, R11
	BR	runtime·morestack(SB)

// reflectcall: call a function with the given argument list
// func call(stackArgsType *_type, f *FuncVal, stackArgs *byte, stackArgsSize, stackRetOffset, frameSize uint32, regArgs *abi.RegArgs).
// we don't have variable-sized frames, so we use a small number
// of constant-sized-frame functions to encode a few bits of size in the pc.

#define DISPATCH(NAME,MAXSIZE)		\
	MOVD	$MAXSIZE, R31;		\
	CMPU	R3, R31;		\
	BGT	4(PC);			\
	MOVD	$NAME(SB), R12;		\
	MOVD	R12, CTR;		\
	BR	(CTR)

TEXT ·reflectcall(SB),NOSPLIT|NOFRAME,$0-28
	MOVWZ	frameSize+20(FP), R3
	DISPATCH(runtime·call16, 16)
	DISPATCH(runtime·call32, 32)
	DISPATCH(runtime·call64, 64)
	DISPATCH(runtime·call128, 128)
	DISPATCH(runtime·call256, 256)
	DISPATCH(runtime·call512, 512)
	DISPATCH(runtime·call1024, 1024)
	DISPATCH(runtime·call2048, 2048)
	DISPATCH(runtime·call4096, 4096)
	DISPATCH(runtime·call8192, 8192)
	DISPATCH(runtime·call16384, 16384)
	DISPATCH(runtime·call32768, 32768)
	DISPATCH(runtime·call65536, 65536)
	DISPATCH(runtime·call131072, 131072)
	DISPATCH(runtime·call262144, 262144)
	DISPATCH(runtime·call524288, 524288)
	DISPATCH(runtime·call1048576, 1048576)
	DISPATCH(runtime·call2097152, 2097152)
	DISPATCH(runtime·call4194304, 4194304)
	DISPATCH(runtime·call8388608, 8388608)
	DISPATCH(runtime·call16777216, 16777216)
	DISPATCH(runtime·call33554432, 33554432)
	DISPATCH(runtime·call67108864, 67108864)
	DISPATCH(runtime·call134217728, 134217728)
	DISPATCH(runtime·call268435456, 268435456)
	DISPATCH(runtime·call536870912, 536870912)
	DISPATCH(runtime·call1073741824, 1073741824)
	MOVD	$runtime·badreflectcall(SB), R12
	MOVD	R12, CTR
	BR	(CTR)

#define CALLFN(NAME,MAXSIZE)			\
TEXT NAME(SB), WRAPPER, $MAXSIZE-28;		\
	NO_LOCAL_POINTERS;			\
	/* copy arguments to stack */		\
	MOVD	stackArgs+8(FP), R3;		\
	MOVWZ	stackArgsSize+12(FP), R4;	\
	MOVD	R1, R5;				\
	CMP	R4, $0;				\
	BEQ	callfn;				\
	ADD	$(FIXED_FRAME-1), R5;		\
	SUB	$1, R3;				\
tail: \
	MOVBZU	1(R3), R6;			\
	MOVBU	R6, 1(R5);			\
	SUB	$1, R4;				\
	CMP	R4, $0;				\
	BGT	tail;				\
callfn: \
	/* call function */			\
	MOVD	f+4(FP), R11;			\
	MOVD	(R11), R12;			\
	MOVD	R12, CTR;			\
	PCDATA	$PCDATA_StackMapIndex, $0;	\
	BL	(CTR);				\
	/* copy return values back */		\
	MOVD	stackArgsType+0(FP), R7;	\
	MOVD	stackArgs+8(FP), R3;		\
	MOVWZ	stackArgsSize+12(FP), R4;	\
	MOVWZ	stackRetOffset+16(FP), R6;	\
	ADD	$FIXED_FRAME, R1, R5;		\
	ADD	R6, R5;				\
	ADD	R6, R3;				\
	SUB	R6, R4;				\
	BL	callRet<>(SB);			\
	RET

// callRet copies return values back at the end of call*. This is a
// separate function so it can allocate stack space for the arguments
// to reflectcallmove. It does not follow the Go ABI; it expects its
// arguments in registers.
TEXT callRet<>(SB), NOSPLIT, $20-0
	MOVW	R7, FIXED_FRAME+0(R1)
	MOVW	R3, FIXED_FRAME+4(R1)
	MOVW	R5, FIXED_FRAME+8(R1)
	MOVW	R4, FIXED_FRAME+12(R1)
	MOVW	R0, FIXED_FRAME+16(R1)
	BL	runtime·reflectcallmove(SB)
	RET

CALLFN(·call16, 16)
CALLFN(·call32, 32)
CALLFN(·call64, 64)
CALLFN(·call128, 128)
CALLFN(·call256, 256)
CALLFN(·call512, 512)
CALLFN(·call1024, 1024)
CALLFN(·call2048, 2048)
CALLFN(·call4096, 4096)
CALLFN(·call8192, 8192)
CALLFN(·call16384, 16384)
CALLFN(·call32768, 32768)
CALLFN(·call65536, 65536)
CALLFN(·call131072, 131072)
CALLFN(·call262144, 262144)
CALLFN(·call524288, 524288)
CALLFN(·call1048576, 1048576)
CALLFN(·call2097152, 2097152)
CALLFN(·call4194304, 4194304)
CALLFN(·call8388608, 8388608)
CALLFN(·call16777216, 16777216)
CALLFN(·call33554432, 33554432)
CALLFN(·call67108864, 67108864)
CALLFN(·call134217728, 134217728)
CALLFN(·call268435456, 268435456)
CALLFN(·call536870912, 536870912)
CALLFN(·call1073741824, 1073741824)

TEXT runtime·procyieldAsm(SB),NOSPLIT|NOFRAME,$0-4
	RET

// Save state of caller into g->sched,
// but using fake PC from systemstack_switch.
// Must only be called from functions with no locals ($0)
// or else unwinding from systemstack_switch is incorrect.
// Smashes R31.
TEXT gosave_systemstack_switch<>(SB),NOSPLIT|NOFRAME,$0
	MOVD	$runtime·systemstack_switch(SB), R31
	ADD	$8, R31	// get past prologue
	MOVD	R31, (g_sched+gobuf_pc)(g)
	MOVD	R1, (g_sched+gobuf_sp)(g)
	MOVD	R0, (g_sched+gobuf_lr)(g)
	// Assert ctxt is zero. See func save.
	MOVD	(g_sched+gobuf_ctxt)(g), R31
	CMP	R31, $0
	BEQ	2(PC)
	BL	runtime·abort(SB)
	RET

// func asmcgocall(fn, arg unsafe.Pointer) int32
// Call fn(arg) on the scheduler stack,
// aligned appropriately for the gcc ABI.
// See cgocall.go for more details.
TEXT ·asmcgocall(SB),NOSPLIT,$0-12
	MOVD	fn+0(FP), R12
	MOVD	arg+4(FP), R3

	MOVD	R1, R7	// save original stack pointer
	MOVD	g, R5

	// Figure out if we need to switch to m->g0 stack.
	// We get called to create new OS threads too, and those
	// come in on the m->g0 stack already. Or we might already
	// be on the m->gsignal stack.
	MOVD	g_m(g), R8
	MOVD	m_gsignal(R8), R6
	CMP	R6, g
	BEQ	g0
	MOVD	m_g0(R8), R6
	CMP	R6, g
	BEQ	g0
	BL	gosave_systemstack_switch<>(SB)
	MOVD	R6, g
	BL	runtime·save_g(SB)
	MOVD	(g_sched+gobuf_sp)(g), R1

	// Now on a scheduling stack (a pthread-created stack).
g0:
	// SysV needs a 16-byte aligned stack and a back chain word at 0(R1).
	SUB	$32, R1
	RLWNM	$0, R1, $~15, R1
	MOVW	R5, 24(R1)	// save old g on stack
	MOVD	(g_stack+stack_hi)(R5), R5
	SUB	R7, R5
	MOVW	R5, 28(R1)	// save depth in old g stack (can't just save SP, as stack might be copied during a callback)
	MOVW	R0, 0(R1)
	MOVD	R12, CTR
	BL	(CTR)
	MOVW	$0, R0	// C code may have clobbered R0

	// Restore g, stack pointer. R3 is return value.
	MOVW	24(R1), g
	BL	runtime·save_g(SB)
	MOVD	(g_stack+stack_hi)(g), R5
	MOVW	28(R1), R6
	SUB	R6, R5
	MOVD	R5, R1

	MOVW	R3, ret+8(FP)
	RET

// cgocallback(fn, frame unsafe.Pointer, ctxt uintptr)
// cgo is not supported on linux/ppc yet.
TEXT ·cgocallback(SB),NOSPLIT,$0-12
	BL	runtime·abort(SB)
	RET

// void setg(G*); set g. for use by needm.
TEXT runtime·setg(SB),NOSPLIT,$0-4
	MOVD	gg+0(FP), g
	// This only happens if iscgo, so jump straight to save_g
	BL	runtime·save_g(SB)
	RET

TEXT runtime·abort(SB),NOSPLIT|NOFRAME,$0-0
	MOVW	(R0), R0
	UNDEF

// The top-most function running on a goroutine
// returns to goexit+PCQuantum.
TEXT runtime·goexit(SB),NOSPLIT|NOFRAME|TOPFRAME,$0-0
	MOVD	R0, R0	// NOP
	BL	runtime·goexit1(SB)	// does not return
	// traceback from goexit1 must hit code range of goexit
	MOVD	R0, R0	// NOP

// gcWriteBarrier informs the GC about heap pointer writes.
//
// gcWriteBarrier does NOT follow the Go ABI. It accepts the
// number of bytes of buffer needed in R29, and returns a pointer
// to the buffer space in R29.
// It clobbers condition codes.
// It does not clobber R0 through R17 (except special registers),
// but may clobber any other register, *including* R31.
TEXT gcWriteBarrier<>(SB),NOSPLIT,$60
	// The standard prologue clobbers R31.
	// We use R18, R19, and R31 as scratch registers.
retry:
	MOVD	g_m(g), R18
	MOVD	m_p(R18), R18
	MOVD	(p_wbBuf+wbBuf_next)(R18), R19
	MOVD	(p_wbBuf+wbBuf_end)(R18), R31
	// Increment wbBuf.next position.
	ADD	R29, R19
	// Is the buffer full?
	CMPU	R31, R19
	BLT	flush
	// Commit to the larger buffer.
	MOVD	R19, (p_wbBuf+wbBuf_next)(R18)
	// Make return value (the original next position)
	SUB	R29, R19, R29
	RET

flush:
	// Save the registers the compiler expects to survive (see LoweredWB).
	MOVD	R20, (FIXED_FRAME+0)(R1)
	MOVD	R21, (FIXED_FRAME+4)(R1)
	MOVD	R3, (FIXED_FRAME+8)(R1)
	MOVD	R4, (FIXED_FRAME+12)(R1)
	MOVD	R5, (FIXED_FRAME+16)(R1)
	MOVD	R6, (FIXED_FRAME+20)(R1)
	MOVD	R7, (FIXED_FRAME+24)(R1)
	MOVD	R8, (FIXED_FRAME+28)(R1)
	MOVD	R9, (FIXED_FRAME+32)(R1)
	MOVD	R10, (FIXED_FRAME+36)(R1)
	MOVD	R14, (FIXED_FRAME+40)(R1)
	MOVD	R15, (FIXED_FRAME+44)(R1)
	MOVD	R16, (FIXED_FRAME+48)(R1)
	MOVD	R17, (FIXED_FRAME+52)(R1)
	MOVD	R29, (FIXED_FRAME+56)(R1)

	CALL	runtime·wbBufFlush(SB)

	MOVD	(FIXED_FRAME+0)(R1), R20
	MOVD	(FIXED_FRAME+4)(R1), R21
	MOVD	(FIXED_FRAME+8)(R1), R3
	MOVD	(FIXED_FRAME+12)(R1), R4
	MOVD	(FIXED_FRAME+16)(R1), R5
	MOVD	(FIXED_FRAME+20)(R1), R6
	MOVD	(FIXED_FRAME+24)(R1), R7
	MOVD	(FIXED_FRAME+28)(R1), R8
	MOVD	(FIXED_FRAME+32)(R1), R9
	MOVD	(FIXED_FRAME+36)(R1), R10
	MOVD	(FIXED_FRAME+40)(R1), R14
	MOVD	(FIXED_FRAME+44)(R1), R15
	MOVD	(FIXED_FRAME+48)(R1), R16
	MOVD	(FIXED_FRAME+52)(R1), R17
	MOVD	(FIXED_FRAME+56)(R1), R29
	JMP	retry

TEXT runtime·gcWriteBarrier1<ABIInternal>(SB),NOSPLIT,$0
	MOVD	$4, R29
	JMP	gcWriteBarrier<>(SB)
TEXT runtime·gcWriteBarrier2<ABIInternal>(SB),NOSPLIT,$0
	MOVD	$8, R29
	JMP	gcWriteBarrier<>(SB)
TEXT runtime·gcWriteBarrier3<ABIInternal>(SB),NOSPLIT,$0
	MOVD	$12, R29
	JMP	gcWriteBarrier<>(SB)
TEXT runtime·gcWriteBarrier4<ABIInternal>(SB),NOSPLIT,$0
	MOVD	$16, R29
	JMP	gcWriteBarrier<>(SB)
TEXT runtime·gcWriteBarrier5<ABIInternal>(SB),NOSPLIT,$0
	MOVD	$20, R29
	JMP	gcWriteBarrier<>(SB)
TEXT runtime·gcWriteBarrier6<ABIInternal>(SB),NOSPLIT,$0
	MOVD	$24, R29
	JMP	gcWriteBarrier<>(SB)
TEXT runtime·gcWriteBarrier7<ABIInternal>(SB),NOSPLIT,$0
	MOVD	$28, R29
	JMP	gcWriteBarrier<>(SB)
TEXT runtime·gcWriteBarrier8<ABIInternal>(SB),NOSPLIT,$0
	MOVD	$32, R29
	JMP	gcWriteBarrier<>(SB)

// The compiler numbers registers from R3 when encoding a bounds failure
// (see cmd/compile/internal/ppc), so the save area starts at R3.
#define SAVE_BOUNDS_REGS \
	MOVW	R3, 16(R1); \
	MOVW	R4, 20(R1); \
	MOVW	R5, 24(R1); \
	MOVW	R6, 28(R1); \
	MOVW	R7, 32(R1); \
	MOVW	R8, 36(R1); \
	MOVW	R9, 40(R1); \
	MOVW	R10, 44(R1); \
	MOVW	R11, 48(R1); \
	MOVW	R12, 52(R1); \
	MOVW	R13, 56(R1); \
	MOVW	R14, 60(R1); \
	MOVW	R15, 64(R1); \
	MOVW	R16, 68(R1); \
	MOVW	R17, 72(R1); \
	MOVW	R18, 76(R1); \
	MOVD	LR, R31; \
	MOVW	R31, FIXED_FRAME+0(R1);	/* PC immediately after call to panicBounds */ \
	ADD	$16, R1, R3;	/* pointer to save area */ \
	MOVW	R3, FIXED_FRAME+4(R1)

TEXT runtime·panicBounds<ABIInternal>(SB),NOSPLIT,$72-0
	NO_LOCAL_POINTERS
	SAVE_BOUNDS_REGS
	CALL	runtime·panicBounds32<ABIInternal>(SB)
	RET

TEXT runtime·panicExtend<ABIInternal>(SB),NOSPLIT,$72-0
	NO_LOCAL_POINTERS
	SAVE_BOUNDS_REGS
	CALL	runtime·panicBounds32X<ABIInternal>(SB)
	RET
