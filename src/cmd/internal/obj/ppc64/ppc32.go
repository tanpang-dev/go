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

// With -dynlink a Go symbol may live in another module (a plugin's host or a
// Go shared library). The 32-bit PLT stubs expect the GOT pointer in r30,
// which is g here, so instead each module keeps its own pointer to every such
// symbol, GOT32.<sym>, which the dynamic linker binds, and reads it with the
// PC-relative sequence above. Calls through it go via LR (or CTR for tail
// calls), clobbering only REGTMP.
func (c *ctxt9) gotSym32(s *obj.LSym) *obj.LSym {
	return c.ctxt.LookupInit("GOT32."+s.Name, func(g *obj.LSym) {
		g.Type = objabi.SNOPTRDATA
		g.Set(obj.AttrDuplicateOK, true)
		g.Set(obj.AttrStatic, true) // never exported, so never bound to another module's copy
		g.Set(obj.AttrLocal, true)
		c.ctxt.Data = append(c.ctxt.Data, g)
		g.WriteAddr(c.ctxt, 0, 4, s, 0)
	})
}

func isExtern32(a *obj.Addr) bool {
	return a.Name == obj.NAME_EXTERN && a.Sym != nil && !a.Sym.Local() && a.Sym.Type != objabi.STLSBSS
}

// expandCalls32 turns BL/BR to a symbol that may be in another module into
// an indirect call through its GOT32 pointer. It runs after preprocess so
// tail calls already have their epilogue and morestack calls are included.
func (c *ctxt9) expandCalls32() {
	for p := c.cursym.Func().Text.Link; p != nil; p = p.Link {
		if (p.As != ABL && p.As != ABR) || p.To.Type != obj.TYPE_BRANCH || p.To.Sym == nil || p.To.Sym.Local() {
			continue
		}
		reg, as := int16(REG_LR), obj.As(obj.ACALL)
		if p.As == ABR {
			reg, as = REG_CTR, ABR
		}
		// Keep p (and its Spadj/Mark) as the branch; load the target in front.
		s, spadj := p.To.Sym, p.Spadj
		q := obj.Appendp(p, c.newprog)
		q.As = AMOVD
		q.From = obj.Addr{Type: obj.TYPE_REG, Reg: REGTMP}
		q.To = obj.Addr{Type: obj.TYPE_REG, Reg: reg}
		r := obj.Appendp(q, c.newprog)
		r.As = as
		r.To = obj.Addr{Type: obj.TYPE_REG, Reg: reg}
		r.Mark = p.Mark
		r.Spadj = spadj
		p.As = AMOVW
		p.From = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: c.gotSym32(s)}
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REGTMP}
		p.Mark = 0
		p.Spadj = 0
		p = r
	}
}

func (c *ctxt9) rewriteToUseGot32(p *obj.Prog) {
	switch p.As {
	case obj.ATEXT, obj.AFUNCDATA, obj.APCDATA, obj.ANOP:
		return
	}
	loadGot := func(p *obj.Prog, s *obj.LSym, reg int16) {
		p.As = AMOVW
		p.From = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: c.gotSym32(s)}
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: reg}
		p.Reg = 0
	}

	if isExtern32(&p.From) && isExtern32(&p.To) {
		c.ctxt.Diag("cannot handle NAME_EXTERN on both sides in %v with -dynlink", p)
		return
	}
	switch {
	case p.From.Type == obj.TYPE_ADDR && isExtern32(&p.From):
		// MOVW $sym+off, Rx -> MOVW GOT32.sym, Rx; ADD $off, Rx
		if p.To.Type != obj.TYPE_REG {
			c.ctxt.Diag("cannot take the address of %v into a non-register with -dynlink", p)
			return
		}
		off, to := p.From.Offset, p.To
		loadGot(p, p.From.Sym, to.Reg)
		if off != 0 {
			q := obj.Appendp(p, c.newprog)
			q.As = AADD
			q.From = obj.Addr{Type: obj.TYPE_CONST, Offset: off}
			q.To = to
		}
	case p.From.Type == obj.TYPE_MEM && isExtern32(&p.From):
		// MOVx sym+off, Ry -> MOVW GOT32.sym, REGTMP; MOVx off(REGTMP), Ry
		q := obj.Appendp(p, c.newprog)
		q.As, q.Reg, q.To = p.As, p.Reg, p.To
		q.From = obj.Addr{Type: obj.TYPE_MEM, Reg: REGTMP, Offset: p.From.Offset}
		loadGot(p, p.From.Sym, REGTMP)
	case p.To.Type == obj.TYPE_MEM && isExtern32(&p.To):
		// MOVx Ry, sym+off -> MOVW GOT32.sym, REGTMP; MOVx Ry, off(REGTMP)
		if p.From.Type == obj.TYPE_REG && p.From.Reg == REGTMP {
			c.ctxt.Diag("cannot store REGTMP to %v with -dynlink", p)
			return
		}
		q := obj.Appendp(p, c.newprog)
		q.As, q.Reg, q.From = p.As, p.Reg, p.From
		q.To = obj.Addr{Type: obj.TYPE_MEM, Reg: REGTMP, Offset: p.To.Offset}
		loadGot(p, p.To.Sym, REGTMP)
	}
}
