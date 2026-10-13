// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ppc is the linker backend for 32-bit big-endian PowerPC.
// Only internal linking of static executables is supported.
package ppc

import (
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/ld"
	"cmd/link/internal/loader"
)

func Init() (*sys.Arch, ld.Arch) {
	theArch := ld.Arch{
		Funcalign:  FuncAlign,
		Maxalign:   MaxAlign,
		Minalign:   MinAlign,
		Dwarfregsp: DWARFREGSP,
		Dwarfreglr: DWARFREGLR,

		Adddynrel:        adddynrel,
		Archinit:         archinit,
		Archreloc:        archreloc,
		Archrelocvariant: archrelocvariant,
		Extreloc:         extreloc,
		Gentext:          gentext,
		Machoreloc1:      machoreloc1,

		ELF: ld.ELFArch{
			Linuxdynld:     "/lib/ld.so.1",
			LinuxdynldMusl: "/lib/ld-musl-powerpc.so.1",

			Freebsddynld:   "XXX",
			Openbsddynld:   "XXX",
			Netbsddynld:    "XXX",
			Dragonflydynld: "XXX",
			Solarisdynld:   "XXX",

			Reloc1:    elfreloc1,
			RelocSize: 12,
			SetupPLT:  elfsetupplt,
		},
	}

	return sys.ArchPPC, theArch
}

func archinit(ctxt *ld.Link) {
	switch ctxt.HeadType {
	default:
		ld.Exitf("unknown -H option: %v", ctxt.HeadType)
	case objabi.Hlinux: /* ppc elf */
		ld.Elfinit(ctxt)
		ld.HEADR = ld.ELFRESERVE
		if *ld.FlagRound == -1 {
			*ld.FlagRound = 0x10000
		}
		if *ld.FlagTextAddr == -1 {
			*ld.FlagTextAddr = ld.Rnd(0x10000, *ld.FlagRound) + int64(ld.HEADR)
		}
	}
}

func adddynrel(target *ld.Target, ldr *loader.Loader, syms *ld.ArchSyms, s loader.Sym, r loader.Reloc, rIdx int) bool {
	ld.Exitf("adddynrel currently unimplemented for ppc")
	return false
}
