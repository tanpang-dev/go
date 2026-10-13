// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

const (
	MaxAlign  = 32 // max data alignment
	MinAlign  = 1  // min data alignment
	FuncAlign = 16
)

/* Used by ../internal/ld/dwarf.go */
const (
	DWARFREGSP = 1
	DWARFREGLR = 65
)
