// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"
#include "funcdata.h"

#define REGCTXT R11

// makeFuncStub is the code half of the function returned by MakeFunc.
// See the comment on the declaration of makeFuncStub in makefunc.go
// for more details.
// No arg size here, runtime pulls arg map out of the func value.
TEXT ·makeFuncStub(SB),(NOSPLIT|WRAPPER),$20
	NO_LOCAL_POINTERS
	MOVW	REGCTXT, 8(R1)
	MOVD	$argframe+0(FP), R3
	MOVW	R3, 12(R1)
	MOVB	R0, 24(R1)
	ADD	$24, R1, R3
	MOVW	R3, 16(R1)
	MOVW	R0, 20(R1)
	BL	·callReflect(SB)
	RET

// methodValueCall is the code half of the function returned by makeMethodValue.
// See the comment on the declaration of methodValueCall in makefunc.go
// for more details.
// No arg size here; runtime pulls arg map out of the func value.
TEXT ·methodValueCall(SB),(NOSPLIT|WRAPPER),$20
	NO_LOCAL_POINTERS
	MOVW	REGCTXT, 8(R1)
	MOVD	$argframe+0(FP), R3
	MOVW	R3, 12(R1)
	MOVB	R0, 24(R1)
	ADD	$24, R1, R3
	MOVW	R3, 16(R1)
	MOVW	R0, 20(R1)
	BL	·callMethod(SB)
	RET
