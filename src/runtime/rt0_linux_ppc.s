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
