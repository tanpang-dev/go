// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc64

import (
	"cmd/internal/obj"
	"cmd/internal/sys"
)

// 32-bit PowerPC (GOARCH=ppc) reuses this backend. The 750CL has none of the
// 64-bit-only instructions, so the doubleword forms that preprocess and
// hand-written assembly use are mapped onto their word equivalents here,
// right before encoding.

func isPPC32(ctxt *obj.Link) bool { return ctxt.Arch.Family == sys.PPC }

// tlsReg is the thread pointer: R13 in the 64-bit ELF ABI, R2 in the 32-bit one.
func (c *ctxt9) tlsReg() uint32 {
	if isPPC32(c.ctxt) {
		return REG_R2
	}
	return REG_R13
}

func isMem(a *obj.Addr) bool { return a.Type == obj.TYPE_MEM }

func isReg(a *obj.Addr) bool { return a.Type == obj.TYPE_REG && a.Reg >= REG_R0 && a.Reg <= REG_R31 }

// fixPPC32 rewrites one instruction into a form that exists on 32-bit PowerPC.
func fixPPC32(ctxt *obj.Link, p *obj.Prog) {
	switch p.As {
	case AMOVD, AMOVW, AMOVWZ:
		switch {
		case isMem(&p.From) && isReg(&p.To):
			// ld/lwa -> lwz
			p.As = AMOVWZ
		case isReg(&p.From) && isMem(&p.To):
			// std -> stw
			p.As = AMOVW
		case isReg(&p.From) && isReg(&p.To):
			// extsw / clrldi are 64-bit only; a word move is just mr.
			p.As = AMOVD
		case p.From.Type == obj.TYPE_CONST && p.From.Name == obj.NAME_NONE && p.From.Reg == 0:
			// Only the low word matters; keep constants in the li/lis+ori range.
			p.From.Offset = int64(int32(p.From.Offset))
			if p.As == AMOVWZ {
				p.As = AMOVW
			}
		}
	case AMOVDU:
		if isReg(&p.From) && isMem(&p.To) {
			p.As = AMOVWU // stdu -> stwu
		} else if isMem(&p.From) {
			p.As = AMOVWZU // ldu -> lwzu
		}
	case ACMP:
		p.As = ACMPW
	case ACMPU:
		p.As = ACMPWU
	}
}
