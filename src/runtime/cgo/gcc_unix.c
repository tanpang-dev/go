// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix

#include "libcgo.h"
#include "libcgo_unix.h"

#if defined(__powerpc__) && !defined(__powerpc64__)
// 32-bit PowerPC: g lives here whenever cgo is used (see runtime/tls_ppc.s).
// Initial-exec keeps it at one offset from the thread pointer (r2) in every
// thread. Defined in this file so that c-archive links always pull it in.
__thread void *x_cgo_tls_g __attribute__((tls_model("initial-exec")));

static void
inittls_ppc(void **tlsg, void **tlsbase)
{
	char *tp;

	(void)tlsbase;
	__asm__("mr %0, 2" : "=r"(tp));
	*tlsg = (void *)((char *)&x_cgo_tls_g - tp);
}

void (*x_cgo_inittls)(void **tlsg, void **tlsbase) = inittls_ppc;
#else
// Platform-specific hooks.
void (*x_cgo_inittls)(void **tlsg, void **tlsbase) __attribute__((weak));
#endif
void (*x_cgo_threadentry_platform)(void) __attribute__((weak));

static void (*setg_gcc)(void*);

// _cgo_set_stacklo sets g->stacklo based on the stack size.
// This is common code called from x_cgo_init, which is itself
// called by rt0_go in the runtime package.
static void
_cgo_set_stacklo(G *g)
{
	uintptr bounds[2];

	x_cgo_getstackbound(bounds);

	g->stacklo = bounds[0];

	// Sanity check the results now, rather than getting a
	// morestack on g0 crash.
	if (g->stacklo >= g->stackhi) {
		fprintf(stderr, "runtime/cgo: bad stack bounds: lo=%p hi=%p\n", (void*)(g->stacklo), (void*)(g->stackhi));
		abort();
	}
}

void
x_cgo_init(G *g, void (*setg)(void*), void **tlsg, void **tlsbase)
{
	setg_gcc = setg;
	_cgo_set_stacklo(g);

	if (x_cgo_inittls) {
		x_cgo_inittls(tlsg, tlsbase);
	}
}

void*
threadentry(void *v)
{
	ThreadStart ts;

	ts = *(ThreadStart*)v;
	_cgo_tsan_acquire();
	free(v);
	_cgo_tsan_release();

	if (x_cgo_threadentry_platform != NULL) {
		x_cgo_threadentry_platform();
	}

	crosscall1(ts.fn, setg_gcc, ts.g);
	return NULL;
}
