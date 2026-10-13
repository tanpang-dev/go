// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ppc is the SSA backend for 32-bit big-endian PowerPC (GOARCH=ppc).
package ppc

import (
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/ssagen"
	"cmd/internal/obj/ppc64"
)

func Init(arch *ssagen.ArchInfo) {
	arch.LinkArch = &ppc64.Linkppc
	arch.REGSP = ppc64.REGSP
	arch.MAXWIDTH = (1 << 31) - 1
	// ponytail: soft float everywhere for now; the 750CL FPU (minus fsqrt and
	// the 64-bit int conversions) can be used later.
	arch.SoftFloat = true
	arch.ZeroRange = zerorange
	arch.Ginsnop = ginsnop
	arch.SSAMarkMoves = func(s *ssagen.State, b *ssa.Block) {}
	arch.SSAGenValue = ssaGenValue
	arch.SSAGenBlock = ssaGenBlock
}
