// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && ppc

//
// System calls and other sys.stuff for 32-bit PowerPC, Linux
//
// SYSCALL $n assembles to "li r0,n; sc; xor r0,r0,r0". Arguments go in
// R3-R8; on error the kernel sets CR0.SO and returns a positive errno in R3.

#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"

#define FIXED_FRAME 8

#define SYS_exit		1
#define SYS_read		3
#define SYS_write		4
#define SYS_open		5
#define SYS_close		6
#define SYS_getpid		20
#define SYS_kill		37
#define SYS_brk			45
#define SYS_mmap		90
#define SYS_munmap		91
#define SYS_setitimer		104
#define SYS_clone		120
#define SYS_sched_yield		158
#define SYS_nanosleep		162
#define SYS_rt_sigreturn	172
#define SYS_rt_sigaction	173
#define SYS_rt_sigprocmask	174
#define SYS_sigaltstack		185
#define SYS_madvise		205
#define SYS_mincore		206
#define SYS_gettid		207
#define SYS_futex		221
#define SYS_sched_getaffinity	223
#define SYS_exit_group		234
#define SYS_timer_create	240
#define SYS_timer_settime	241
#define SYS_timer_delete	244
#define SYS_clock_gettime	246
#define SYS_tgkill		250
#define SYS_pipe2		317
#define SYS_clock_gettime64	403
#define SYS_timer_settime64	409
#define SYS_futex_time64	422

TEXT runtime·exit(SB),NOSPLIT|NOFRAME,$0-4
	MOVW	code+0(FP), R3
	SYSCALL	$SYS_exit_group
	RET

// func exitThread(wait *atomic.Uint32)
TEXT runtime·exitThread(SB),NOSPLIT|NOFRAME,$0-4
	MOVW	wait+0(FP), R5
	// We're done using the stack.
	SYNC
	MOVW	R0, 0(R5)
	SYNC
	MOVW	$0, R3	// exit code
	SYSCALL	$SYS_exit
	JMP	0(PC)

TEXT runtime·open(SB),NOSPLIT|NOFRAME,$0-16
	MOVW	name+0(FP), R3
	MOVW	mode+4(FP), R4
	MOVW	perm+8(FP), R5
	SYSCALL	$SYS_open
	BVC	2(PC)
	MOVW	$-1, R3
	MOVW	R3, ret+12(FP)
	RET

TEXT runtime·closefd(SB),NOSPLIT|NOFRAME,$0-8
	MOVW	fd+0(FP), R3
	SYSCALL	$SYS_close
	BVC	2(PC)
	MOVW	$-1, R3
	MOVW	R3, ret+4(FP)
	RET

TEXT runtime·write1(SB),NOSPLIT|NOFRAME,$0-16
	MOVW	fd+0(FP), R3
	MOVW	p+4(FP), R4
	MOVW	n+8(FP), R5
	SYSCALL	$SYS_write
	BVC	2(PC)
	NEG	R3	// caller expects negative errno
	MOVW	R3, ret+12(FP)
	RET

TEXT runtime·read(SB),NOSPLIT|NOFRAME,$0-16
	MOVW	fd+0(FP), R3
	MOVW	p+4(FP), R4
	MOVW	n+8(FP), R5
	SYSCALL	$SYS_read
	BVC	2(PC)
	NEG	R3	// caller expects negative errno
	MOVW	R3, ret+12(FP)
	RET

// func pipe2(flags int32) (r, w int32, errno int32)
TEXT runtime·pipe2(SB),NOSPLIT|NOFRAME,$0-16
	MOVD	$r+4(FP), R3
	MOVW	flags+0(FP), R4
	SYSCALL	$SYS_pipe2
	BVC	2(PC)
	NEG	R3	// caller expects negative errno
	MOVW	R3, errno+12(FP)
	RET

TEXT runtime·usleep(SB),NOSPLIT,$16-4
	MOVW	usec+0(FP), R3
	MOVW	$1000000, R4
	DIVWU	R4, R3, R5	// R5 = sec
	MULLW	R4, R5, R6
	SUB	R6, R3, R6	// R6 = usec % 1e6
	MOVW	$1000, R4
	MULLW	R4, R6	// nsec
	MOVW	R5, FIXED_FRAME+0(R1)
	MOVW	R6, FIXED_FRAME+4(R1)

	// nanosleep(&ts, 0)
	ADD	$FIXED_FRAME, R1, R3
	MOVW	$0, R4
	SYSCALL	$SYS_nanosleep
	RET

TEXT runtime·gettid(SB),NOSPLIT|NOFRAME,$0-4
	SYSCALL	$SYS_gettid
	MOVW	R3, ret+0(FP)
	RET

TEXT runtime·raise(SB),NOSPLIT|NOFRAME,$0-4
	SYSCALL	$SYS_getpid
	MOVW	R3, R14
	SYSCALL	$SYS_gettid
	MOVW	R3, R4	// arg 2 tid
	MOVW	R14, R3	// arg 1 pid
	MOVW	sig+0(FP), R5	// arg 3
	SYSCALL	$SYS_tgkill
	RET

TEXT runtime·raiseproc(SB),NOSPLIT|NOFRAME,$0
	SYSCALL	$SYS_getpid
	MOVW	R3, R3	// arg 1 pid
	MOVW	sig+0(FP), R4	// arg 2
	SYSCALL	$SYS_kill
	RET

TEXT ·getpid(SB),NOSPLIT|NOFRAME,$0-4
	SYSCALL	$SYS_getpid
	MOVW	R3, ret+0(FP)
	RET

TEXT ·tgkill(SB),NOSPLIT|NOFRAME,$0-12
	MOVW	tgid+0(FP), R3
	MOVW	tid+4(FP), R4
	MOVW	sig+8(FP), R5
	SYSCALL	$SYS_tgkill
	RET

TEXT runtime·setitimer(SB),NOSPLIT|NOFRAME,$0-12
	MOVW	mode+0(FP), R3
	MOVW	new+4(FP), R4
	MOVW	old+8(FP), R5
	SYSCALL	$SYS_setitimer
	RET

TEXT runtime·timer_create(SB),NOSPLIT|NOFRAME,$0-16
	MOVW	clockid+0(FP), R3
	MOVW	sevp+4(FP), R4
	MOVW	timerid+8(FP), R5
	SYSCALL	$SYS_timer_create
	BVC	2(PC)
	NEG	R3
	MOVW	R3, ret+12(FP)
	RET

TEXT runtime·timer_settime32(SB),NOSPLIT|NOFRAME,$0-20
	MOVW	timerid+0(FP), R3
	MOVW	flags+4(FP), R4
	MOVW	new+8(FP), R5
	MOVW	old+12(FP), R6
	SYSCALL	$SYS_timer_settime
	BVC	2(PC)
	NEG	R3
	MOVW	R3, ret+16(FP)
	RET

TEXT runtime·timer_settime64(SB),NOSPLIT|NOFRAME,$0-20
	MOVW	timerid+0(FP), R3
	MOVW	flags+4(FP), R4
	MOVW	new+8(FP), R5
	MOVW	old+12(FP), R6
	SYSCALL	$SYS_timer_settime64
	BVC	2(PC)
	NEG	R3
	MOVW	R3, ret+16(FP)
	RET

TEXT runtime·timer_delete(SB),NOSPLIT|NOFRAME,$0-8
	MOVW	timerid+0(FP), R3
	SYSCALL	$SYS_timer_delete
	BVC	2(PC)
	NEG	R3
	MOVW	R3, ret+4(FP)
	RET

TEXT runtime·mincore(SB),NOSPLIT|NOFRAME,$0-16
	MOVW	addr+0(FP), R3
	MOVW	n+4(FP), R4
	MOVW	dst+8(FP), R5
	SYSCALL	$SYS_mincore
	BVC	2(PC)
	NEG	R3	// caller expects negative errno
	MOVW	R3, ret+12(FP)
	RET

// vdsoCall calls the VDSO function in R12 with the C arguments in R3 and R4
// and leaves its result in R3. Like the arm version it switches to the g0
// stack (the VDSO's stack needs are unknown), records vdsoPC/vdsoSP for
// SIGPROF tracebacks and, without cgo, parks g at the bottom of the signal
// stack so sigFetchG can find it if a signal lands inside the VDSO.
// Clobbers R15-R18 (callee-saved in C, so they survive the call).
TEXT vdsoCall<>(SB),NOSPLIT,$16-0
	MOVD	R1, R15	// our SP
	MOVD	g_m(g), R16

	// Save the old vdsoPC/vdsoSP so this is reentrant.
	MOVD	m_vdsoPC(R16), R17
	MOVW	R17, FIXED_FRAME+0(R1)
	MOVD	m_vdsoSP(R16), R17
	MOVW	R17, FIXED_FRAME+4(R1)
	MOVD	LR, R17
	MOVD	R17, m_vdsoPC(R16)
	ADD	$(16+FIXED_FRAME), R1, R17	// caller's SP
	MOVD	R17, m_vdsoSP(R16)

	MOVD	m_curg(R16), R17
	CMP	g, R17	// Only switch if on curg.
	BNE	noswitch
	MOVD	m_g0(R16), R17
	MOVD	(g_sched+gobuf_sp)(R17), R1
noswitch:
	SUB	$32, R1
	RLWNM	$0, R1, $~15, R1	// SysV wants 16-byte alignment
	MOVW	R0, 0(R1)	// back chain

	MOVD	R0, R18
	MOVBZ	runtime·iscgo(SB), R17
	CMP	R17, $0
	BNE	call
	MOVD	m_gsignal(R16), R17
	CMP	R17, $0
	BEQ	call
	CMP	g, R17	// already on the signal stack: no nested signal
	BEQ	call
	MOVD	(g_stack+stack_lo)(R17), R18
	CMP	R18, $0
	BEQ	call
	MOVD	g, 0(R18)
call:
	MOVD	R12, CTR
	BL	(CTR)
	MOVW	$0, R0	// C code may clobber R0
	CMP	R18, $0
	BEQ	2(PC)
	MOVW	R0, 0(R18)	// clear the g slot

	MOVD	R15, R1
	MOVW	FIXED_FRAME+4(R1), R17
	MOVD	R17, m_vdsoSP(R16)
	MOVW	FIXED_FRAME+0(R1), R17
	MOVD	R17, m_vdsoPC(R16)
	RET

// clock_gettime64 / __kernel_clock_gettime64 fill a __kernel_timespec:
// two big-endian int64s.

// func walltime() (sec int64, nsec int32)
TEXT runtime·walltime(SB),NOSPLIT,$16-12
	MOVW	$0, R3	// CLOCK_REALTIME
	ADD	$FIXED_FRAME, R1, R4
	MOVW	runtime·vdsoClockgettimeSym(SB), R12
	CMP	R12, $0
	BEQ	wallsys
	BL	vdsoCall<>(SB)
	BR	wallok
wallsys:
	SYSCALL	$SYS_clock_gettime64
wallok:
	MOVW	FIXED_FRAME+0(R1), R5	// sec high
	MOVW	FIXED_FRAME+4(R1), R6	// sec low
	MOVW	FIXED_FRAME+12(R1), R7	// nsec (low word)
	MOVW	R5, sec_hi+0(FP)
	MOVW	R6, sec_lo+4(FP)
	MOVW	R7, nsec+8(FP)
	RET

// func nanotime1() int64
TEXT runtime·nanotime1(SB),NOSPLIT,$16-8
	MOVW	$1, R3	// CLOCK_MONOTONIC
	ADD	$FIXED_FRAME, R1, R4
	MOVW	runtime·vdsoClockgettimeSym(SB), R12
	CMP	R12, $0
	BEQ	nanosys
	BL	vdsoCall<>(SB)
	BR	nanook
nanosys:
	SYSCALL	$SYS_clock_gettime64
nanook:
	MOVW	FIXED_FRAME+0(R1), R5	// sec high
	MOVW	FIXED_FRAME+4(R1), R6	// sec low
	MOVW	FIXED_FRAME+12(R1), R7	// nsec
	MOVW	$1000000000, R8
	MULLW	R8, R6, R9	// low word of sec*1e9
	MULHWU	R8, R6, R10	// high word of sec_low*1e9
	MULLW	R8, R5, R11	// sec_high*1e9 contributes to the high word only
	ADD	R11, R10
	ADDC	R7, R9	// + nsec
	ADDZE	R10
	MOVW	R10, ret_hi+0(FP)
	MOVW	R9, ret_lo+4(FP)
	RET

TEXT runtime·rtsigprocmask(SB),NOSPLIT|NOFRAME,$0-16
	MOVW	how+0(FP), R3
	MOVW	new+4(FP), R4
	MOVW	old+8(FP), R5
	MOVW	size+12(FP), R6
	SYSCALL	$SYS_rt_sigprocmask
	BVC	2(PC)
	MOVW	R0, 0xf0(R0)	// crash
	RET

TEXT runtime·rt_sigaction(SB),NOSPLIT|NOFRAME,$0-20
	MOVW	sig+0(FP), R3
	MOVW	new+4(FP), R4
	MOVW	old+8(FP), R5
	MOVW	size+12(FP), R6
	SYSCALL	$SYS_rt_sigaction
	BVC	2(PC)
	NEG	R3	// caller expects negative errno
	MOVW	R3, ret+16(FP)
	RET

// func sigfwd(fn uintptr, sig uint32, info *siginfo, ctx unsafe.Pointer)
// Calls fn with the C calling convention on a 16-byte aligned stack.
TEXT runtime·sigfwd(SB),NOSPLIT,$0-16
	MOVW	sig+4(FP), R3
	MOVW	info+8(FP), R4
	MOVW	ctx+12(FP), R5
	MOVW	fn+0(FP), R12
	MOVD	R1, R14
	SUB	$16, R1
	RLWNM	$0, R1, $~15, R1
	MOVW	R14, 0(R1)	// back chain
	MOVD	R12, CTR
	BL	(CTR)
	MOVD	R14, R1
	MOVW	$0, R0
	RET

// Called by the kernel with R3=sig, R4=info, R5=ctx and LR pointing at the
// vdso sigreturn trampoline. The kernel restores every register from the
// signal frame on return, so nothing needs to be saved here.
TEXT runtime·sigtramp(SB),NOSPLIT|TOPFRAME,$16
	MOVW	$0, R0
	// The signal may arrive in C code, where R30 is not g.
	MOVBZ	runtime·iscgo(SB), R6
	CMP	R6, $0
	BEQ	2(PC)
	BL	runtime·load_g(SB)
	MOVW	R3, FIXED_FRAME+0(R1)
	MOVW	R4, FIXED_FRAME+4(R1)
	MOVW	R5, FIXED_FRAME+8(R1)
	MOVD	$runtime·sigtrampgo(SB), R12
	MOVD	R12, CTR
	BL	(CTR)
	RET

TEXT runtime·cgoSigtramp(SB),NOSPLIT|NOFRAME,$0
	JMP	runtime·sigtramp(SB)

TEXT runtime·mmap(SB),NOSPLIT|NOFRAME,$0-32
	MOVW	addr+0(FP), R3
	MOVW	n+4(FP), R4
	MOVW	prot+8(FP), R5
	MOVW	flags+12(FP), R6
	MOVW	fd+16(FP), R7
	MOVW	off+20(FP), R8

	SYSCALL	$SYS_mmap
	BVC	ok
	MOVW	R0, p+24(FP)
	MOVW	R3, err+28(FP)
	RET
ok:
	MOVW	R3, p+24(FP)
	MOVW	R0, err+28(FP)
	RET

TEXT runtime·munmap(SB),NOSPLIT|NOFRAME,$0-8
	MOVW	addr+0(FP), R3
	MOVW	n+4(FP), R4
	SYSCALL	$SYS_munmap
	BVC	2(PC)
	MOVW	R0, 0xf0(R0)	// crash
	RET

TEXT runtime·madvise(SB),NOSPLIT|NOFRAME,$0-16
	MOVW	addr+0(FP), R3
	MOVW	n+4(FP), R4
	MOVW	flags+8(FP), R5
	SYSCALL	$SYS_madvise
	BVC	2(PC)
	NEG	R3
	MOVW	R3, ret+12(FP)
	RET

// int32 futex(int32 *uaddr, int32 op, int32 val,
//	struct timespec *timeout, int32 *uaddr2, int32 val2);
TEXT runtime·futex_time32(SB),NOSPLIT|NOFRAME,$0-28
	MOVW	addr+0(FP), R3
	MOVW	op+4(FP), R4
	MOVW	val+8(FP), R5
	MOVW	ts+12(FP), R6
	MOVW	addr2+16(FP), R7
	MOVW	val3+20(FP), R8
	SYSCALL	$SYS_futex
	BVC	2(PC)
	NEG	R3	// caller expects negative errno
	MOVW	R3, ret+24(FP)
	RET

TEXT runtime·futex_time64(SB),NOSPLIT|NOFRAME,$0-28
	MOVW	addr+0(FP), R3
	MOVW	op+4(FP), R4
	MOVW	val+8(FP), R5
	MOVW	ts+12(FP), R6
	MOVW	addr2+16(FP), R7
	MOVW	val3+20(FP), R8
	SYSCALL	$SYS_futex_time64
	BVC	2(PC)
	NEG	R3	// caller expects negative errno
	MOVW	R3, ret+24(FP)
	RET

// int32 clone(int32 flags, void *stk, M *mp, G *gp, void (*fn)(void));
TEXT runtime·clone(SB),NOSPLIT|NOFRAME,$0-24
	MOVW	flags+0(FP), R3
	MOVW	stk+4(FP), R4

	// Copy mp, gp, fn off parent stack for use by child.
	MOVW	mp+8(FP), R7
	MOVW	gp+12(FP), R8
	MOVW	fn+16(FP), R12

	MOVW	R7, -4(R4)
	MOVW	R8, -8(R4)
	MOVW	R12, -12(R4)
	MOVW	$1234, R7
	MOVW	R7, -16(R4)
	SUB	$32, R4	// leave room for the words above; keep 16-byte alignment

	MOVW	$0, R5	// ptid
	MOVW	$0, R6	// tls
	MOVW	$0, R7	// ctid
	SYSCALL	$SYS_clone
	BVC	2(PC)
	NEG	R3	// caller expects negative errno

	// In parent, return.
	CMP	R3, $0
	BEQ	3(PC)
	MOVW	R3, ret+20(FP)
	RET

	// In child, on new stack.
	ADD	$32, R1
	MOVW	-16(R1), R7
	CMP	R7, $1234
	BEQ	2(PC)
	MOVW	R0, 0(R0)

	// Initialize m->procid to Linux tid
	SYSCALL	$SYS_gettid

	MOVW	-12(R1), R12	// fn
	MOVW	-8(R1), R8	// g
	MOVW	-4(R1), R7	// m

	CMP	R7, $0
	BEQ	nog
	CMP	R8, $0
	BEQ	nog

	MOVW	R3, m_procid(R7)

	// In child, set up new stack
	MOVW	R7, g_m(R8)
	MOVD	R8, g

nog:
	SUB	$32, R1
	// Call fn
	MOVD	R12, CTR
	BL	(CTR)

	// It shouldn't return.	 If it does, exit that thread.
	MOVW	$111, R3
	SYSCALL	$SYS_exit
	BR	-2(PC)	// keep exiting

TEXT runtime·sigaltstack(SB),NOSPLIT|NOFRAME,$0
	MOVW	new+0(FP), R3
	MOVW	old+4(FP), R4
	SYSCALL	$SYS_sigaltstack
	BVC	2(PC)
	MOVW	R0, 0xf0(R0)	// crash
	RET

TEXT runtime·osyield(SB),NOSPLIT|NOFRAME,$0
	SYSCALL	$SYS_sched_yield
	RET

TEXT runtime·sched_getaffinity(SB),NOSPLIT|NOFRAME,$0-16
	MOVW	pid+0(FP), R3
	MOVW	len+4(FP), R4
	MOVW	buf+8(FP), R5
	SYSCALL	$SYS_sched_getaffinity
	BVC	2(PC)
	NEG	R3	// caller expects negative errno
	MOVW	R3, ret+12(FP)
	RET

// func sbrk0() uintptr
TEXT runtime·sbrk0(SB),NOSPLIT|NOFRAME,$0-4
	// Implemented as brk(NULL).
	MOVW	$0, R3
	SYSCALL	$SYS_brk
	MOVW	R3, ret+0(FP)
	RET

TEXT runtime·access(SB),NOSPLIT|NOFRAME,$0-12
	MOVW	R0, 0(R0) // unimplemented, only needed for android; declared in stubs_linux.go
	MOVW	R0, ret+8(FP)	// for vet
	RET

TEXT runtime·connect(SB),NOSPLIT|NOFRAME,$0-16
	MOVW	R0, 0(R0) // unimplemented, only needed for android; declared in stubs_linux.go
	MOVW	R0, ret+12(FP)	// for vet
	RET

TEXT runtime·socket(SB),NOSPLIT|NOFRAME,$0-16
	MOVW	R0, 0(R0) // unimplemented, only needed for android; declared in stubs_linux.go
	MOVW	R0, ret+12(FP)	// for vet
	RET
