// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "funcdata.h"
#include "textflag.h"

// We have to resort to TLS variable to save g (R30).
// One reason is that external code might trigger
// SIGSEGV, and our runtime.sigtramp don't even know we
// are in external code, and will continue to use R30,
// this might well result in another SIGSEGV.
// g lives in runtime/cgo's initial-exec __thread slot x_cgo_tls_g
// (gcc_unix.c), whose offset from R2 (the 32-bit SysV thread pointer)
// x_cgo_inittls stores in tls_g_off. Every build mode uses the same slot, so
// a plugin's copy of the runtime agrees with its host's. Without cgo,
// save_g/load_g are never needed.
#define TLS_G(r)	MOVW	runtime·tls_g_off(SB), r; ADD	R2, r

// If !iscgo, this is a no-op.
//
// NOTE: setg_gcc<> assume this clobbers only R31.
TEXT runtime·save_g(SB),NOSPLIT|NOFRAME,$0-0
	MOVBZ	runtime·iscgo(SB), R31
	CMP	R31, $0
	BEQ	nocgo
	TLS_G(R31)
	MOVD	g, 0(R31)

nocgo:
	RET

// load_g loads the g register from thread-local memory, for use after
// calling externally compiled code that overwrote those registers.
//
// NOTE: _cgo_topofstack assumes this only clobbers g (R30), and R31.
TEXT runtime·load_g(SB),NOSPLIT|NOFRAME,$0-0
	TLS_G(R31)
	MOVD	0(R31), g
	RET

GLOBL runtime·tls_g_off(SB), NOPTR, $4
