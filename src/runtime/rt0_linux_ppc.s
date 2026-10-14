// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT _rt0_ppc_linux(SB),NOSPLIT|NOFRAME,$0
	// In a statically linked binary, the stack contains argc,
	// argv as argc string pointers followed by a NULL, envv as a
	// sequence of string pointers followed by a NULL, and auxv.
	MOVW	$0, R0
	MOVW	0(R1), R3	// argc
	ADD	$4, R1, R4	// argv
	BR	main(SB)

TEXT main(SB),NOSPLIT|NOFRAME,$0
	MOVD	$runtime·rt0_go(SB), R12
	MOVD	R12, CTR
	BR	(CTR)

// _rt0_ppc_linux_lib is the entry for -buildmode=c-archive and c-shared,
// run as an ELF constructor with argc and argv in R3 and R4 (SysV ABI).
// R14-R31, F14-F31 and CR are callee-saved there.
// Frame: 0 back chain, 24-95 R14-R31, 96 CR, 104-247 F14-F31.
TEXT _rt0_ppc_linux_lib(SB),NOSPLIT|NOFRAME,$0
	MOVW	LR, R0
	MOVW	R0, 4(R1)	// LR save word of the caller's frame
	MOVW	R1, R0
	SUB	$256, R1
	MOVW	R0, 0(R1)	// back chain
	MOVW	R14, 24(R1)
	MOVW	R15, 28(R1)
	MOVW	R16, 32(R1)
	MOVW	R17, 36(R1)
	MOVW	R18, 40(R1)
	MOVW	R19, 44(R1)
	MOVW	R20, 48(R1)
	MOVW	R21, 52(R1)
	MOVW	R22, 56(R1)
	MOVW	R23, 60(R1)
	MOVW	R24, 64(R1)
	MOVW	R25, 68(R1)
	MOVW	R26, 72(R1)
	MOVW	R27, 76(R1)
	MOVW	R28, 80(R1)
	MOVW	R29, 84(R1)
	MOVW	g, 88(R1)
	MOVW	R31, 92(R1)
	FMOVD	F14, 104(R1)
	FMOVD	F15, 112(R1)
	FMOVD	F16, 120(R1)
	FMOVD	F17, 128(R1)
	FMOVD	F18, 136(R1)
	FMOVD	F19, 144(R1)
	FMOVD	F20, 152(R1)
	FMOVD	F21, 160(R1)
	FMOVD	F22, 168(R1)
	FMOVD	F23, 176(R1)
	FMOVD	F24, 184(R1)
	FMOVD	F25, 192(R1)
	FMOVD	F26, 200(R1)
	FMOVD	F27, 208(R1)
	FMOVD	F28, 216(R1)
	FMOVD	F29, 224(R1)
	FMOVD	F30, 232(R1)
	FMOVD	F31, 240(R1)
	MOVW	CR, R0
	MOVW	R0, 96(R1)
	MOVW	$0, R0

	MOVW	R3, _rt0_ppc_lib_argc<>(SB)
	MOVW	R4, _rt0_ppc_lib_argv<>(SB)
	MOVW	$0, g
	BL	runtime·libInit(SB)

	MOVW	96(R1), R0
	MOVW	R0, CR
	MOVW	24(R1), R14
	MOVW	28(R1), R15
	MOVW	32(R1), R16
	MOVW	36(R1), R17
	MOVW	40(R1), R18
	MOVW	44(R1), R19
	MOVW	48(R1), R20
	MOVW	52(R1), R21
	MOVW	56(R1), R22
	MOVW	60(R1), R23
	MOVW	64(R1), R24
	MOVW	68(R1), R25
	MOVW	72(R1), R26
	MOVW	76(R1), R27
	MOVW	80(R1), R28
	MOVW	84(R1), R29
	MOVW	88(R1), g
	MOVW	92(R1), R31
	FMOVD	104(R1), F14
	FMOVD	112(R1), F15
	FMOVD	120(R1), F16
	FMOVD	128(R1), F17
	FMOVD	136(R1), F18
	FMOVD	144(R1), F19
	FMOVD	152(R1), F20
	FMOVD	160(R1), F21
	FMOVD	168(R1), F22
	FMOVD	176(R1), F23
	FMOVD	184(R1), F24
	FMOVD	192(R1), F25
	FMOVD	200(R1), F26
	FMOVD	208(R1), F27
	FMOVD	216(R1), F28
	FMOVD	224(R1), F29
	FMOVD	232(R1), F30
	FMOVD	240(R1), F31
	ADD	$256, R1
	MOVW	4(R1), R0
	MOVW	R0, LR
	RET

// rt0_lib_go initializes the Go runtime.
// This is started in a separate thread by _rt0_ppc_linux_lib.
TEXT runtime·rt0_lib_go<ABIInternal>(SB),NOSPLIT|NOFRAME,$0
	MOVW	$0, R0
	MOVW	_rt0_ppc_lib_argc<>(SB), R3
	MOVW	_rt0_ppc_lib_argv<>(SB), R4
	MOVD	$runtime·rt0_go(SB), R12
	MOVD	R12, CTR
	BR	(CTR)

DATA _rt0_ppc_lib_argc<>(SB)/4, $0
GLOBL _rt0_ppc_lib_argc<>(SB), NOPTR, $4
DATA _rt0_ppc_lib_argv<>(SB)/4, $0
GLOBL _rt0_ppc_lib_argv<>(SB), NOPTR, $4
