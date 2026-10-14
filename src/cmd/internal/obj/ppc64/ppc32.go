// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc64

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
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
	if p.From.Sym != nil && p.From.Sym.Type == objabi.STLSBSS {
		return // MOVD tlsvar(SB), R computes the variable's address
	}
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

// Position-independent code on 32-bit PowerPC. There is no PC-relative
// addressing, so with -shared every symbol access fetches the PC first:
//
//	mflr r0; bcl 20,31,.+4; mflr r31; mtlr r0; li r0,0
//	addis rX, r31, (sym - base)@ha; op rY, (sym - base)@l(rX)
//
// base is the address after the bcl. LR is preserved and R0 put back to
// zero, so the sequence works in leaf and NOFRAME functions too, and being
// a single Prog that uses REGTMP it is never an async preemption point.
const pcrelPrefixSize = 20

// pcrelBaseOff is the distance from base to the addis.
const pcrelBaseOff = pcrelPrefixSize - 8

func isSymbolAccess(o Optab) bool {
	return o.type_ == 74 || o.type_ == 75 || o.type_ == 26 && o.a1 == C_LACON
}

func symAddr(a *obj.Addr) bool {
	return a.Sym != nil && (a.Name == obj.NAME_EXTERN || a.Name == obj.NAME_STATIC)
}

func (c *ctxt9) symbolAccessPC32(s *obj.LSym, d int64, reg int16, op uint32, reuse bool) (o1, o2 uint32, rel obj.Reloc) {
	base := uint32(REGTMP)
	if !reuse && reg == REGTMP {
		c.ctxt.Diag("cannot store REGTMP to a symbol with -shared: %v", s.Name)
	}
	if reuse {
		o1 = AOP_IRR(OP_ADDIS, uint32(reg), base, 0)
		o2 = AOP_IRR(op, uint32(reg), uint32(reg), 0)
	} else {
		o1 = AOP_IRR(OP_ADDIS, REGTMP, base, 0)
		o2 = AOP_IRR(op, uint32(reg), REGTMP, 0)
	}
	rel = obj.Reloc{
		Type: objabi.R_ADDRPOWER_PCREL,
		Off:  int32(c.pc) + pcrelPrefixSize,
		Siz:  8,
		Sym:  s,
		Add:  d + pcrelBaseOff,
	}
	return
}

// prefixPC32 moves the instruction words of p behind the PC fetch, or behind
// NOPs when this form is used without a symbol (large offsets off a register).
func prefixPC32(p *obj.Prog, out *[8]uint32) {
	copy(out[pcrelPrefixSize/4:], out[:len(out)-pcrelPrefixSize/4])
	if symAddr(&p.From) || symAddr(&p.To) {
		copy(out[:], []uint32{0x7c0802a6, 0x429f0005, 0x7fe802a6, 0x7c0803a6, 0x38000000})
	} else {
		copy(out[:], []uint32{NOP, NOP, NOP, NOP, NOP})
	}
}
