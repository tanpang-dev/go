// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && ppc

package runtime

import (
	"internal/goarch"
	"unsafe"
)

// Indexes into mcontext.gregs (asm/ptrace.h PT_*).
const (
	_PT_NIP  = 32
	_PT_CTR  = 35
	_PT_LINK = 36
	_PT_XER  = 37
	_PT_CCR  = 38
	_PT_TRAP = 40
	_PT_DAR  = 41
)

type sigctxt struct {
	info *siginfo
	ctxt unsafe.Pointer
}

//go:nosplit
//go:nowritebarrierrec
func (c *sigctxt) regs() *[48]uint32 { return &(*ucontext)(c.ctxt).uc_regs.gregs }

func (c *sigctxt) r0() uint32  { return c.regs()[0] }
func (c *sigctxt) r1() uint32  { return c.regs()[1] }
func (c *sigctxt) r2() uint32  { return c.regs()[2] }
func (c *sigctxt) r3() uint32  { return c.regs()[3] }
func (c *sigctxt) r4() uint32  { return c.regs()[4] }
func (c *sigctxt) r5() uint32  { return c.regs()[5] }
func (c *sigctxt) r6() uint32  { return c.regs()[6] }
func (c *sigctxt) r7() uint32  { return c.regs()[7] }
func (c *sigctxt) r8() uint32  { return c.regs()[8] }
func (c *sigctxt) r9() uint32  { return c.regs()[9] }
func (c *sigctxt) r10() uint32 { return c.regs()[10] }
func (c *sigctxt) r11() uint32 { return c.regs()[11] }
func (c *sigctxt) r12() uint32 { return c.regs()[12] }
func (c *sigctxt) r13() uint32 { return c.regs()[13] }
func (c *sigctxt) r14() uint32 { return c.regs()[14] }
func (c *sigctxt) r15() uint32 { return c.regs()[15] }
func (c *sigctxt) r16() uint32 { return c.regs()[16] }
func (c *sigctxt) r17() uint32 { return c.regs()[17] }
func (c *sigctxt) r18() uint32 { return c.regs()[18] }
func (c *sigctxt) r19() uint32 { return c.regs()[19] }
func (c *sigctxt) r20() uint32 { return c.regs()[20] }
func (c *sigctxt) r21() uint32 { return c.regs()[21] }
func (c *sigctxt) r22() uint32 { return c.regs()[22] }
func (c *sigctxt) r23() uint32 { return c.regs()[23] }
func (c *sigctxt) r24() uint32 { return c.regs()[24] }
func (c *sigctxt) r25() uint32 { return c.regs()[25] }
func (c *sigctxt) r26() uint32 { return c.regs()[26] }
func (c *sigctxt) r27() uint32 { return c.regs()[27] }
func (c *sigctxt) r28() uint32 { return c.regs()[28] }
func (c *sigctxt) r29() uint32 { return c.regs()[29] }
func (c *sigctxt) r30() uint32 { return c.regs()[30] }
func (c *sigctxt) r31() uint32 { return c.regs()[31] }
func (c *sigctxt) sp() uint32  { return c.regs()[1] }

//go:nosplit
//go:nowritebarrierrec
func (c *sigctxt) pc() uint32 { return c.regs()[_PT_NIP] }

func (c *sigctxt) trap() uint32 { return c.regs()[_PT_TRAP] }
func (c *sigctxt) ctr() uint32  { return c.regs()[_PT_CTR] }
func (c *sigctxt) link() uint32 { return c.regs()[_PT_LINK] }
func (c *sigctxt) xer() uint32  { return c.regs()[_PT_XER] }
func (c *sigctxt) ccr() uint32  { return c.regs()[_PT_CCR] }

func (c *sigctxt) sigcode() uint32 { return uint32(c.info.si_code) }
func (c *sigctxt) sigaddr() uint32 { return c.info.si_addr }
func (c *sigctxt) fault() uintptr  { return uintptr(c.regs()[_PT_DAR]) }

func (c *sigctxt) set_r0(x uint32)   { c.regs()[0] = x }
func (c *sigctxt) set_r12(x uint32)  { c.regs()[12] = x }
func (c *sigctxt) set_r30(x uint32)  { c.regs()[30] = x }
func (c *sigctxt) set_pc(x uint32)   { c.regs()[_PT_NIP] = x }
func (c *sigctxt) set_sp(x uint32)   { c.regs()[1] = x }
func (c *sigctxt) set_link(x uint32) { c.regs()[_PT_LINK] = x }

func (c *sigctxt) set_sigcode(x uint32) { c.info.si_code = int32(x) }
func (c *sigctxt) set_sigaddr(x uint32) {
	*(*uintptr)(add(unsafe.Pointer(c.info), 2*goarch.PtrSize)) = uintptr(x)
}
