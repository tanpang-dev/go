// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"debug/elf"

	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/ld"
	"cmd/link/internal/loader"
	"cmd/link/internal/sym"
)

func gentext(ctxt *ld.Link, ldr *loader.Loader) {}

// debug/elf does not name these (binutils include/elf/ppc.h).
const (
	rPPC_REL16_LO elf.R_PPC = 250
	rPPC_REL16_HA elf.R_PPC = 252
)

func elfreloc1(ctxt *ld.Link, out *ld.OutBuf, ldr *loader.Loader, s loader.Sym, r loader.ExtReloc, ri int, sectoff int64) bool {
	elfsym := ld.ElfSymForReloc(ctxt, r.Xsym)
	rela := func(t elf.R_PPC, off, add int64) {
		out.Write32(uint32(off))
		out.Write32(uint32(elfsym)<<8 | uint32(t))
		out.Write32(uint32(add))
	}
	// The 16-bit immediates of a big-endian instruction pair sit at +2 and +6.
	switch r.Type {
	default:
		return false
	case objabi.R_ADDR, objabi.R_DWARFSECREF:
		if r.Size != 4 {
			return false
		}
		rela(elf.R_PPC_ADDR32, sectoff, r.Xadd)
	case objabi.R_ADDRPOWER, objabi.R_ADDRPOWER_DS:
		rela(elf.R_PPC_ADDR16_HA, sectoff+2, r.Xadd)
		rela(elf.R_PPC_ADDR16_LO, sectoff+6, r.Xadd)
	case objabi.R_ADDRPOWER_PCREL:
		// P is the halfword being patched, so move the addend with it.
		rela(rPPC_REL16_HA, sectoff+2, r.Xadd+2)
		rela(rPPC_REL16_LO, sectoff+6, r.Xadd+6)
	case objabi.R_POWER_TLS_LE:
		rela(elf.R_PPC_TPREL16_HA, sectoff+2, r.Xadd)
		rela(elf.R_PPC_TPREL16_LO, sectoff+6, r.Xadd)
	case objabi.R_CALLPOWER:
		if r.Size != 4 {
			return false
		}
		rela(elf.R_PPC_REL24, sectoff, r.Xadd)
	}
	return true
}

func elfsetupplt(ctxt *ld.Link, ldr *loader.Loader, plt, gotplt *loader.SymbolBuilder, dynamic loader.Sym) {}

func machoreloc1(*sys.Arch, *ld.OutBuf, *loader.Loader, loader.Sym, loader.ExtReloc, int64) bool {
	return false
}

// computeHA/computeLO split a 32-bit value for an addis (@ha) + D-form (@l) pair.
func computeHA(val int32) uint32 { return uint32(uint16((val + 0x8000) >> 16)) }
func computeLO(val int32) uint32 { return uint32(uint16(val)) }

// The pair is big-endian: the first instruction is the high word of val.
func unpack(val int64) (uint32, uint32) { return uint32(val >> 32), uint32(val) }
func pack(o1, o2 uint32) int64          { return int64(o1)<<32 | int64(o2) }

func archreloc(target *ld.Target, ldr *loader.Loader, syms *ld.ArchSyms, r loader.Reloc, s loader.Sym, val int64) (int64, int, bool) {
	if target.IsExternal() {
		// RELA: the addend lives in the relocation, leave the instructions alone.
		switch r.Type() {
		case objabi.R_ADDRPOWER, objabi.R_ADDRPOWER_DS, objabi.R_ADDRPOWER_PCREL, objabi.R_POWER_TLS_LE:
			return val, 2, true
		case objabi.R_CALLPOWER:
			return val, 1, true
		}
		return val, 0, false
	}
	rs := r.Sym()
	switch r.Type() {
	case objabi.R_ADDRPOWER, objabi.R_ADDRPOWER_DS:
		t := ldr.SymAddr(rs) + r.Add()
		o1, o2 := unpack(val)
		o1 = o1&^0xffff | computeHA(int32(t))
		o2 = o2&^0xffff | computeLO(int32(t))
		return pack(o1, o2), 0, true

	case objabi.R_ADDRPOWER_PCREL:
		// Relative to the addis, the first instruction of the pair.
		t := ldr.SymValue(rs) + r.Add() - (ldr.SymValue(s) + int64(r.Off()))
		o1, o2 := unpack(val)
		o1 = o1&^0xffff | computeHA(int32(t))
		o2 = o2&^0xffff | computeLO(int32(t))
		return pack(o1, o2), 0, true

	case objabi.R_CALLPOWER:
		// Bits 6 through 29 = (S + A - P) >> 2
		t := ldr.SymValue(rs) + r.Add() - (ldr.SymValue(s) + int64(r.Off()))
		if t&3 != 0 {
			ldr.Errorf(s, "relocation for %s+%d is not aligned: %d", ldr.SymName(rs), r.Off(), t)
		}
		if int64(int32(t<<6)>>6) != t {
			ldr.Errorf(s, "direct call too far: %s %x", ldr.SymName(rs), t)
		}
		return val | int64(uint32(t)&^0xfc000003), 0, true

	case objabi.R_POWER_TLS_LE:
		// The thread pointer (R2) points 0x7000 bytes past the start of
		// the TLS block (SysV ABI PowerPC supplement, "TLS").
		v := ldr.SymValue(rs) - 0x7000
		if int64(int32(v)) != v {
			ldr.Errorf(s, "TLS offset out of range %d", v)
		}
		o1, o2 := unpack(val)
		o1 |= computeHA(int32(v))
		o2 |= computeLO(int32(v))
		return pack(o1, o2), 0, true
	}
	return val, 0, false
}

func archrelocvariant(*ld.Target, *loader.Loader, loader.Reloc, sym.RelocVariant, loader.Sym, int64, []byte) int64 {
	return -1
}

func extreloc(target *ld.Target, ldr *loader.Loader, r loader.Reloc, s loader.Sym) (loader.ExtReloc, bool) {
	switch r.Type() {
	case objabi.R_POWER_TLS_LE, objabi.R_CALLPOWER:
		return ld.ExtrelocSimple(ldr, r), true
	case objabi.R_ADDRPOWER, objabi.R_ADDRPOWER_DS, objabi.R_ADDRPOWER_PCREL:
		return ld.ExtrelocViaOuterSym(ldr, r, s), true
	}
	return loader.ExtReloc{}, false
}
