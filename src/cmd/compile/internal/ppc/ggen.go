// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/objw"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/obj/ppc64"
)

func zerorange(pp *objw.Progs, p *obj.Prog, off, cnt int64, _ *uint32) *obj.Prog {
	if cnt%int64(types.PtrSize) != 0 {
		panic("zeroed region not aligned")
	}
	for cnt != 0 {
		p = pp.Append(p, ppc64.AMOVW, obj.TYPE_REG, ppc64.REGZERO, 0, obj.TYPE_MEM, ppc64.REGSP, base.Ctxt.Arch.FixedFrameSize+off)
		cnt -= int64(types.PtrSize)
		off += int64(types.PtrSize)
	}
	return p
}

func ginsnop(pp *objw.Progs) *obj.Prog {
	// Generate the preferred hardware nop: ori 0,0,0
	p := pp.Prog(ppc64.AOR)
	p.From.Type = obj.TYPE_CONST
	p.To.Type = obj.TYPE_REG
	p.To.Reg = ppc64.REG_R0
	return p
}
