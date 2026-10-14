// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

const (
	// vdsoArrayMax is the byte-size of a maximally sized array on this architecture.
	// It must stay below cmd/compile/internal/ppc/galign.go arch.MAXWIDTH
	// ((1<<31)-1); the VDSO is a few pages, so 1<<30 is plenty.
	vdsoArrayMax = 1 << 30
)

var vdsoLinuxVersion = vdsoVersionKey{"LINUX_2.6.15", 0x75fcba5}

var vdsoSymbolKeys = []vdsoSymbolKey{
	// The 64-bit-time variant fills a __kernel_timespec, like clock_gettime64.
	{"__kernel_clock_gettime64", 0xcd72394, 0x6f01bc67, &vdsoClockgettimeSym},
}

// initialize to fall back to syscall
var vdsoClockgettimeSym uintptr = 0
