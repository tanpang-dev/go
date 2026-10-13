// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/logopt"
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/ssa/block"
	"cmd/compile/internal/ssa/ssaop"
	"cmd/compile/internal/ssagen"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/obj/ppc64"
	"internal/abi"
)

// The PPC ops mirror MIPS's (see _gen/PPCOps.go): booleans are 0/1 in a
// register and blocks branch on a register compared with zero. Each op is
// expanded here into a PowerPC sequence. R0 is always zero.

func loadByType(t *types.Type) obj.As {
	switch t.Size() {
	case 1:
		if t.IsSigned() {
			return ppc64.AMOVB
		}
		return ppc64.AMOVBZ
	case 2:
		if t.IsSigned() {
			return ppc64.AMOVH
		}
		return ppc64.AMOVHZ
	case 4:
		return ppc64.AMOVWZ
	}
	panic("bad load type")
}

func storeByType(t *types.Type) obj.As {
	switch t.Size() {
	case 1:
		return ppc64.AMOVB
	case 2:
		return ppc64.AMOVH
	case 4:
		return ppc64.AMOVW
	}
	panic("bad store type")
}

func reg(p *obj.Addr, r int16) {
	p.Type = obj.TYPE_REG
	p.Reg = r
}

// rr emits "as from, [r,] to".
func rr(s *ssagen.State, as obj.As, from, r, to int16) *obj.Prog {
	p := s.Prog(as)
	reg(&p.From, from)
	p.Reg = r
	reg(&p.To, to)
	return p
}

// rc emits "as $c, [r,] to".
func rc(s *ssagen.State, as obj.As, c int64, r, to int16) *obj.Prog {
	p := s.Prog(as)
	p.From.SetConst(c)
	p.Reg = r
	reg(&p.To, to)
	return p
}

func movreg(s *ssagen.State, from, to int16) {
	if from != to {
		rr(s, ppc64.AMOVD, from, 0, to)
	}
}

// crBit copies one bit of CR0 (0=LT, 1=GT, 2=EQ) into the low bit of to.
func crBit(s *ssagen.State, bit int64, to int16) {
	p := s.Prog(ppc64.AMOVW)
	reg(&p.From, ppc64.REG_CR)
	reg(&p.To, to)
	p = s.Prog(ppc64.ARLWNM)
	p.From.SetConst(bit + 1)
	p.Reg = to
	p.AddRestSourceConst(1)
	reg(&p.To, to)
}

func fitsS16(c int64) bool { return c == int64(int16(c)) }
func fitsU16(c int64) bool { return c == int64(uint16(c)) }

// cmpConst compares x against c, signed or unsigned, leaving the result in CR0.
func cmpConst(s *ssagen.State, unsigned bool, x int16, c int64) {
	as := ppc64.ACMPW
	fits := fitsS16(c)
	if unsigned {
		as = ppc64.ACMPWU
		fits = fitsU16(c)
		c = int64(uint32(c))
	} else {
		c = int64(int32(c))
	}
	if fits {
		p := s.Prog(as)
		reg(&p.From, x)
		p.To.SetConst(c)
		return
	}
	rc(s, ppc64.AMOVW, c, 0, ppc64.REGTMP)
	p := s.Prog(as)
	reg(&p.From, x)
	reg(&p.To, ppc64.REGTMP)
}

func cmpReg(s *ssagen.State, unsigned bool, x, y int16) {
	as := ppc64.ACMPW
	if unsigned {
		as = ppc64.ACMPWU
	}
	p := s.Prog(as)
	reg(&p.From, x)
	reg(&p.To, y)
}

func ssaGenValue(s *ssagen.State, v *ssa.Value) {
	switch v.Op {
	case ssaop.OpCopy, ssaop.OpPPCMOVWreg:
		if v.Type.IsMemory() {
			return
		}
		movreg(s, v.Args[0].Reg(), v.Reg())
	case ssaop.OpPPCMOVWnop:
		// nothing to do
	case ssaop.OpLoadReg:
		if v.Type.IsFlags() {
			v.Fatalf("load flags not implemented: %v", v.LongString())
		}
		p := s.Prog(loadByType(v.Type))
		ssagen.AddrAuto(&p.From, v.Args[0])
		reg(&p.To, v.Reg())
	case ssaop.OpStoreReg:
		if v.Type.IsFlags() {
			v.Fatalf("store flags not implemented: %v", v.LongString())
		}
		p := s.Prog(storeByType(v.Type))
		reg(&p.From, v.Args[0].Reg())
		ssagen.AddrAuto(&p.To, v)

	case ssaop.OpPPCADD, ssaop.OpPPCSUB, ssaop.OpPPCAND, ssaop.OpPPCOR, ssaop.OpPPCXOR,
		ssaop.OpPPCNOR, ssaop.OpPPCMUL:
		// to = arg0 op arg1 (SUB: to = Reg - From)
		rr(s, v.Op.Asm(), v.Args[1].Reg(), v.Args[0].Reg(), v.Reg())
	case ssaop.OpPPCSLL, ssaop.OpPPCSRL, ssaop.OpPPCSRA:
		// The rules only use these with amounts below 32, where PowerPC's
		// 6-bit shift amount agrees with MIPS's mod-32 one.
		rr(s, v.Op.Asm(), v.Args[1].Reg(), v.Args[0].Reg(), v.Reg())
	case ssaop.OpPPCADDconst:
		rc(s, ppc64.AADD, v.AuxInt, v.Args[0].Reg(), v.Reg())
	case ssaop.OpPPCSUBconst:
		rc(s, ppc64.AADD, int64(-int32(v.AuxInt)), v.Args[0].Reg(), v.Reg())
	case ssaop.OpPPCANDconst, ssaop.OpPPCORconst, ssaop.OpPPCXORconst:
		c := int64(uint32(v.AuxInt))
		if c > 0xffff && v.Op != ssaop.OpPPCANDconst && c&0xffff == 0 {
			rc(s, v.Op.Asm(), c, v.Args[0].Reg(), v.Reg())
			break
		}
		if fitsU16(c) {
			as := v.Op.Asm()
			if v.Op == ssaop.OpPPCANDconst {
				as = ppc64.AANDCC // andi. is the only immediate AND
			}
			rc(s, as, c, v.Args[0].Reg(), v.Reg())
			break
		}
		rc(s, ppc64.AMOVW, int64(int32(c)), 0, ppc64.REGTMP)
		rr(s, v.Op.Asm(), ppc64.REGTMP, v.Args[0].Reg(), v.Reg())
	case ssaop.OpPPCSLLconst, ssaop.OpPPCSRLconst, ssaop.OpPPCSRAconst:
		rc(s, v.Op.Asm(), v.AuxInt&31, v.Args[0].Reg(), v.Reg())

	case ssaop.OpPPCMULT, ssaop.OpPPCMULTU:
		// results (hi, lo); outputs never overlap the inputs.
		hi := ppc64.AMULHW
		if v.Op == ssaop.OpPPCMULTU {
			hi = ppc64.AMULHWU
		}
		x, y := v.Args[0].Reg(), v.Args[1].Reg()
		rr(s, hi, y, x, v.Reg0())
		rr(s, ppc64.AMULLW, y, x, v.Reg1())
	case ssaop.OpPPCDIV, ssaop.OpPPCDIVU:
		// results (arg0 % arg1, arg0 / arg1); outputs never overlap the inputs.
		// Division by zero is checked by the caller. divw is undefined for
		// MinInt32 / -1, where Go wants quotient MinInt32 and remainder 0.
		x, y := v.Args[0].Reg(), v.Args[1].Reg()
		rem, quo := v.Reg0(), v.Reg1()
		var skip *obj.Prog
		if v.Op == ssaop.OpPPCDIV {
			cmpConst(s, false, y, -1)
			bne := s.Prog(ppc64.ABNE)
			bne.To.Type = obj.TYPE_BRANCH
			rr(s, ppc64.ANEG, x, 0, quo)
			rr(s, ppc64.AMOVD, ppc64.REGZERO, 0, rem)
			skip = s.Prog(ppc64.ABR)
			skip.To.Type = obj.TYPE_BRANCH
			bne.To.SetTarget(rr(s, ppc64.ADIVW, y, x, quo))
		} else {
			rr(s, ppc64.ADIVWU, y, x, quo)
		}
		rr(s, ppc64.AMULLW, y, quo, ppc64.REGTMP)
		rr(s, ppc64.ASUB, ppc64.REGTMP, x, rem)
		if skip != nil {
			skip.To.SetTarget(s.Prog(obj.ANOP))
		}

	case ssaop.OpPPCSGT, ssaop.OpPPCSGTU:
		// arg0 > arg1
		cmpReg(s, v.Op == ssaop.OpPPCSGTU, v.Args[0].Reg(), v.Args[1].Reg())
		crBit(s, 1, v.Reg())
	case ssaop.OpPPCSGTconst, ssaop.OpPPCSGTUconst:
		// auxInt > arg0, i.e. arg0 < auxInt
		cmpConst(s, v.Op == ssaop.OpPPCSGTUconst, v.Args[0].Reg(), v.AuxInt)
		crBit(s, 0, v.Reg())
	case ssaop.OpPPCSGTzero, ssaop.OpPPCSGTUzero:
		// arg0 > 0
		cmpConst(s, v.Op == ssaop.OpPPCSGTUzero, v.Args[0].Reg(), 0)
		crBit(s, 1, v.Reg())

	case ssaop.OpPPCMOVWconst:
		rc(s, ppc64.AMOVW, int64(int32(v.AuxInt)), 0, v.Reg())
	case ssaop.OpPPCCMOVZ:
		// out (= arg0) = arg2 == 0 ? arg1 : arg0
		cmpConst(s, false, v.Args[2].Reg(), 0)
		bne := s.Prog(ppc64.ABNE)
		bne.To.Type = obj.TYPE_BRANCH
		rr(s, ppc64.AMOVD, v.Args[1].Reg(), 0, v.Reg())
		bne.To.SetTarget(s.Prog(obj.ANOP))
	case ssaop.OpPPCCMOVZzero:
		// out (= arg0) = arg1 == 0 ? 0 : arg0
		cmpConst(s, false, v.Args[1].Reg(), 0)
		bne := s.Prog(ppc64.ABNE)
		bne.To.Type = obj.TYPE_BRANCH
		rr(s, ppc64.AMOVD, ppc64.REGZERO, 0, v.Reg())
		bne.To.SetTarget(s.Prog(obj.ANOP))

	case ssaop.OpPPCMOVWaddr:
		p := s.Prog(ppc64.AMOVD)
		p.From.Type = obj.TYPE_ADDR
		p.From.Reg = v.Args[0].Reg()
		var wantreg string
		switch v.Aux.(type) {
		default:
			v.Fatalf("aux is of unknown type %T", v.Aux)
		case *obj.LSym:
			wantreg = "SB"
			ssagen.AddAux(&p.From, v)
		case *ir.Name:
			wantreg = "SP"
			ssagen.AddAux(&p.From, v)
		case nil:
			wantreg = "SP"
			p.From.Offset = v.AuxInt
		}
		if r := v.Args[0].RegName(); r != wantreg {
			v.Fatalf("bad reg %s for symbol type %T, want %s", r, v.Aux, wantreg)
		}
		reg(&p.To, v.Reg())
	case ssaop.OpPPCMOVBload, ssaop.OpPPCMOVBUload, ssaop.OpPPCMOVHload,
		ssaop.OpPPCMOVHUload, ssaop.OpPPCMOVWload:
		p := s.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_MEM
		p.From.Reg = v.Args[0].Reg()
		ssagen.AddAux(&p.From, v)
		reg(&p.To, v.Reg())
	case ssaop.OpPPCMOVBstore, ssaop.OpPPCMOVHstore, ssaop.OpPPCMOVWstore:
		p := s.Prog(v.Op.Asm())
		reg(&p.From, v.Args[1].Reg())
		p.To.Type = obj.TYPE_MEM
		p.To.Reg = v.Args[0].Reg()
		ssagen.AddAux(&p.To, v)
	case ssaop.OpPPCMOVBstorezero, ssaop.OpPPCMOVHstorezero, ssaop.OpPPCMOVWstorezero:
		p := s.Prog(v.Op.Asm())
		reg(&p.From, ppc64.REGZERO)
		p.To.Type = obj.TYPE_MEM
		p.To.Reg = v.Args[0].Reg()
		ssagen.AddAux(&p.To, v)
	case ssaop.OpPPCMOVBreg, ssaop.OpPPCMOVBUreg, ssaop.OpPPCMOVHreg, ssaop.OpPPCMOVHUreg:
		a := v.Args[0]
		for a.Op == ssaop.OpCopy || a.Op == ssaop.OpPPCMOVWreg || a.Op == ssaop.OpPPCMOVWnop {
			a = a.Args[0]
		}
		if a.Op == ssaop.OpLoadReg {
			t := a.Type
			switch {
			case v.Op == ssaop.OpPPCMOVBreg && t.Size() == 1 && t.IsSigned(),
				v.Op == ssaop.OpPPCMOVBUreg && t.Size() == 1 && !t.IsSigned(),
				v.Op == ssaop.OpPPCMOVHreg && t.Size() == 2 && t.IsSigned(),
				v.Op == ssaop.OpPPCMOVHUreg && t.Size() == 2 && !t.IsSigned():
				// arg is a proper-typed load, already zero/sign-extended
				movreg(s, v.Args[0].Reg(), v.Reg())
				return
			}
		}
		rr(s, v.Op.Asm(), v.Args[0].Reg(), 0, v.Reg())
	case ssaop.OpPPCCLZ:
		rr(s, ppc64.ACNTLZW, v.Args[0].Reg(), 0, v.Reg())
	case ssaop.OpPPCNEG:
		rr(s, ppc64.ANEG, v.Args[0].Reg(), 0, v.Reg())

	case ssaop.OpPPCLoweredZero:
		// arg0 is the first address (R3, advanced), arg1 the last element.
		sz, mov := moveSize(v.AuxInt)
		rc(s, ppc64.AADD, -sz, ppc64.REG_R3, ppc64.REG_R3)
		p2 := s.Prog(mov)
		reg(&p2.From, ppc64.REGZERO)
		p2.To.Type = obj.TYPE_MEM
		p2.To.Reg = ppc64.REG_R3
		p2.To.Offset = sz
		rc(s, ppc64.AADD, sz, ppc64.REG_R3, ppc64.REG_R3)
		cmpReg(s, true, ppc64.REG_R3, v.Args[1].Reg())
		p4 := s.Prog(ppc64.ABNE)
		p4.To.Type = obj.TYPE_BRANCH
		p4.To.SetTarget(p2)
	case ssaop.OpPPCLoweredMove:
		// arg0 = dst (R3), arg1 = src (R4), arg2 = last element of src.
		sz, mov := moveSize(v.AuxInt)
		ld := mov
		if mov == ppc64.AMOVW {
			ld = ppc64.AMOVWZ
		}
		rc(s, ppc64.AADD, -sz, ppc64.REG_R4, ppc64.REG_R4)
		p2 := s.Prog(ld)
		p2.From.Type = obj.TYPE_MEM
		p2.From.Reg = ppc64.REG_R4
		p2.From.Offset = sz
		reg(&p2.To, ppc64.REGTMP)
		p3 := s.Prog(mov)
		reg(&p3.From, ppc64.REGTMP)
		p3.To.Type = obj.TYPE_MEM
		p3.To.Reg = ppc64.REG_R3
		rc(s, ppc64.AADD, sz, ppc64.REG_R4, ppc64.REG_R4)
		rc(s, ppc64.AADD, sz, ppc64.REG_R3, ppc64.REG_R3)
		cmpReg(s, true, ppc64.REG_R4, v.Args[2].Reg())
		p6 := s.Prog(ppc64.ABNE)
		p6.To.Type = obj.TYPE_BRANCH
		p6.To.SetTarget(p2)

	case ssaop.OpPPCCALLstatic:
		s.Call(v)
	case ssaop.OpPPCCALLclosure, ssaop.OpPPCCALLinter:
		rr(s, ppc64.AMOVD, v.Args[0].Reg(), 0, ppc64.REG_CTR)
		pp := s.Call(v)
		pp.To.Reg = ppc64.REG_CTR
	case ssaop.OpPPCCALLtail, ssaop.OpPPCCALLtailinter:
		// obj9 moves a register target into CTR for the tail branch.
		s.TailCall(v)
	case ssaop.OpPPCLoweredWB:
		p := s.Prog(obj.ACALL)
		p.To.Type = obj.TYPE_MEM
		p.To.Name = obj.NAME_EXTERN
		p.To.Sym = ir.Syms.GCWriteBarrier[v.AuxInt-1]

	case ssaop.OpPPCLoweredPanicBoundsRR, ssaop.OpPPCLoweredPanicBoundsRC, ssaop.OpPPCLoweredPanicBoundsCR,
		ssaop.OpPPCLoweredPanicBoundsCC, ssaop.OpPPCLoweredPanicExtendRR, ssaop.OpPPCLoweredPanicExtendRC:
		panicBounds(s, v)

	case ssaop.OpPPCLoweredNilCheck:
		p := s.Prog(ppc64.AMOVBZ)
		p.From.Type = obj.TYPE_MEM
		p.From.Reg = v.Args[0].Reg()
		ssagen.AddAux(&p.From, v)
		reg(&p.To, ppc64.REGTMP)
		if logopt.Enabled() {
			logopt.LogOpt(v.Pos, "nilcheck", "genssa", v.Block.Func.Name)
		}
		if base.Debug.Nil != 0 && v.Pos.Line() > 1 { // v.Pos.Line()==1 in generated wrappers
			base.WarnfAt(v.Pos, "generated nil check")
		}
	case ssaop.OpPPCLoweredGetClosurePtr:
		ssagen.CheckLoweredGetClosurePtr(v)
	case ssaop.OpPPCLoweredGetCallerSP:
		p := s.Prog(ppc64.AMOVD)
		p.From.Type = obj.TYPE_ADDR
		p.From.Offset = -base.Ctxt.Arch.FixedFrameSize
		p.From.Name = obj.NAME_PARAM
		reg(&p.To, v.Reg())
	case ssaop.OpPPCLoweredGetCallerPC:
		p := s.Prog(obj.AGETCALLERPC)
		reg(&p.To, v.Reg())
	case ssaop.OpPPCLoweredPubBarrier:
		s.Prog(v.Op.Asm())
	case ssaop.OpClobber, ssaop.OpClobberReg:
		// TODO: implement for clobberdead experiment. Nop is ok for now.
	default:
		v.Fatalf("genValue not implemented: %s", v.LongString())
	}
}

func moveSize(align int64) (int64, obj.As) {
	switch {
	case align%4 == 0:
		return 4, ppc64.AMOVW
	case align%2 == 0:
		return 2, ppc64.AMOVH
	}
	return 1, ppc64.AMOVB
}

// panicBounds follows mips: register operands are numbered from R3.
func panicBounds(s *ssagen.State, v *ssa.Value) {
	code, signed := ssa.BoundsKind(v.AuxInt).Code()
	xIsReg, yIsReg := false, false
	xVal, yVal := 0, 0
	extend := false
	r := func(n int) int16 { return ppc64.REG_R3 + int16(n) }
	num := func(rg int16) int { return int(rg - ppc64.REG_R3) }
	setConst := func(n int, c int64) { rc(s, ppc64.AMOVW, int64(int32(c)), 0, r(n)) }
	switch v.Op {
	case ssaop.OpPPCLoweredPanicBoundsRR:
		xIsReg, xVal = true, num(v.Args[0].Reg())
		yIsReg, yVal = true, num(v.Args[1].Reg())
	case ssaop.OpPPCLoweredPanicExtendRR:
		extend = true
		hi, lo := num(v.Args[0].Reg()), num(v.Args[1].Reg())
		xIsReg, xVal = true, hi<<2+lo // encode 2 register numbers
		yIsReg, yVal = true, num(v.Args[2].Reg())
	case ssaop.OpPPCLoweredPanicBoundsRC:
		xIsReg, xVal = true, num(v.Args[0].Reg())
		c := v.Aux.(ssa.PanicBoundsC).C
		if c >= 0 && c <= abi.BoundsMaxConst {
			yVal = int(c)
		} else {
			yIsReg = true
			if yVal == xVal {
				yVal = 1
			}
			setConst(yVal, c)
		}
	case ssaop.OpPPCLoweredPanicExtendRC:
		extend = true
		hi, lo := num(v.Args[0].Reg()), num(v.Args[1].Reg())
		xIsReg, xVal = true, hi<<2+lo
		c := v.Aux.(ssa.PanicBoundsC).C
		if c >= 0 && c <= abi.BoundsMaxConst {
			yVal = int(c)
		} else {
			for yVal == hi || yVal == lo {
				yVal++
			}
			yIsReg = true
			setConst(yVal, c)
		}
	case ssaop.OpPPCLoweredPanicBoundsCR:
		yIsReg, yVal = true, num(v.Args[0].Reg())
		c := v.Aux.(ssa.PanicBoundsC).C
		switch {
		case c >= 0 && c <= abi.BoundsMaxConst:
			xVal = int(c)
		case signed && int64(int32(c)) == c || !signed && int64(uint32(c)) == c:
			xIsReg = true
			if xVal == yVal {
				xVal = 1
			}
			setConst(xVal, c)
		default:
			extend, xIsReg = true, true
			hi, lo := 0, 1
			if hi == yVal {
				hi = 2
			}
			if lo == yVal {
				lo = 2
			}
			xVal = hi<<2 + lo
			setConst(hi, c>>32)
			setConst(lo, c)
		}
	case ssaop.OpPPCLoweredPanicBoundsCC:
		c := v.Aux.(ssa.PanicBoundsCC).Cx
		switch {
		case c >= 0 && c <= abi.BoundsMaxConst:
			xVal = int(c)
		case signed && int64(int32(c)) == c || !signed && int64(uint32(c)) == c:
			xIsReg = true
			setConst(xVal, c)
		default:
			extend, xIsReg = true, true
			xVal = 0<<2 + 1
			setConst(0, c>>32)
			setConst(1, c)
		}
		c = v.Aux.(ssa.PanicBoundsCC).Cy
		if c >= 0 && c <= abi.BoundsMaxConst {
			yVal = int(c)
		} else {
			yIsReg, yVal = true, 2
			setConst(yVal, c)
		}
	}
	c := abi.BoundsEncode(code, signed, xIsReg, yIsReg, xVal, yVal)

	p := s.Prog(obj.APCDATA)
	p.From.SetConst(abi.PCDATA_PanicBounds)
	p.To.SetConst(int64(c))
	p = s.Prog(obj.ACALL)
	p.To.Type = obj.TYPE_MEM
	p.To.Name = obj.NAME_EXTERN
	if extend {
		p.To.Sym = ir.Syms.PanicExtend
	} else {
		p.To.Sym = ir.Syms.PanicBounds
	}
}

// Blocks compare their control register with zero.
var blockJump = map[block.BlockKind]struct {
	asm, invasm obj.As
}{
	block.BlockPPCEQ:  {ppc64.ABEQ, ppc64.ABNE},
	block.BlockPPCNE:  {ppc64.ABNE, ppc64.ABEQ},
	block.BlockPPCLTZ: {ppc64.ABLT, ppc64.ABGE},
	block.BlockPPCGEZ: {ppc64.ABGE, ppc64.ABLT},
	block.BlockPPCLEZ: {ppc64.ABLE, ppc64.ABGT},
	block.BlockPPCGTZ: {ppc64.ABGT, ppc64.ABLE},
}

func ssaGenBlock(s *ssagen.State, b, next *ssa.Block) {
	switch b.Kind {
	case block.BlockPlain, block.BlockDefer:
		if b.Succs[0].Block() != next {
			p := s.Prog(obj.AJMP)
			p.To.Type = obj.TYPE_BRANCH
			s.Branches = append(s.Branches, ssagen.Branch{P: p, B: b.Succs[0].Block()})
		}
	case block.BlockExit, block.BlockRetJmp:
	case block.BlockRet:
		s.Prog(obj.ARET)
	case block.BlockPPCEQ, block.BlockPPCNE,
		block.BlockPPCLTZ, block.BlockPPCGEZ,
		block.BlockPPCLEZ, block.BlockPPCGTZ:
		jmp := blockJump[b.Kind]
		cmpConst(s, false, b.Controls[0].Reg(), 0)
		switch next {
		case b.Succs[0].Block():
			s.Br(jmp.invasm, b.Succs[1].Block())
		case b.Succs[1].Block():
			s.Br(jmp.asm, b.Succs[0].Block())
		default:
			if b.Likely != ssa.BranchUnlikely {
				s.Br(jmp.asm, b.Succs[0].Block())
				s.Br(obj.AJMP, b.Succs[1].Block())
			} else {
				s.Br(jmp.invasm, b.Succs[1].Block())
				s.Br(obj.AJMP, b.Succs[0].Block())
			}
		}
	case block.BlockPPCJUMPTABLE:
		// SLW $2, idx, Rtmp; ADD tbl, Rtmp; MOVWZ (Rtmp), Rtmp; MOVD Rtmp, CTR; BR (CTR)
		rc(s, ppc64.ASLW, 2, b.Controls[0].Reg(), ppc64.REGTMP)
		rr(s, ppc64.AADD, b.Controls[1].Reg(), ppc64.REGTMP, ppc64.REGTMP)
		p := s.Prog(ppc64.AMOVWZ)
		p.From.Type = obj.TYPE_MEM
		p.From.Reg = ppc64.REGTMP
		reg(&p.To, ppc64.REGTMP)
		rr(s, ppc64.AMOVD, ppc64.REGTMP, 0, ppc64.REG_CTR)
		p = s.Prog(obj.AJMP)
		reg(&p.To, ppc64.REG_CTR)
		// Save jump tables for later resolution of the target blocks.
		s.JumpTables = append(s.JumpTables, b)
	default:
		b.Fatalf("branch not implemented: %s", b.LongString())
	}
}
