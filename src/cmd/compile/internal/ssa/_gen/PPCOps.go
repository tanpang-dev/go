// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "strings"

// Notes:
//  - Integer types live in the low portion of registers. Upper portions are junk.
//  - Boolean types use the low-order byte of a register. 0=false, 1=true.
//    Upper bytes are junk.
//  - Unused portions of AuxInt are filled by sign-extending the used portion.
//  - *const instructions may use a constant larger than the instruction can encode.
//    In this case the assembler expands to multiple instructions and uses tmp
//    register (R31).
//  - Ops follow MIPSOps.go so that PPC.rules can stay close to MIPS.rules;
//    ppc/ssa.go turns them into PowerPC instruction sequences.
//  - R0 is kept zero, R1 is SP, R2 is the thread pointer, R13 is reserved
//    (SysV small data), R30 is g and R31 is REGTMP.

// Suffixes encode the bit width of various instructions.
// W (word)      = 32 bit
// H (half word) = 16 bit
// HU            = 16 bit unsigned
// B (byte)      = 8 bit
// BU            = 8 bit unsigned
// F (float)     = 32 bit float
// D (double)    = 64 bit float

// Note: registers not used in regalloc are not included in this list,
// so that regmask stays within int64
// Be careful when hand coding regmasks.
var regNamesPPC = []string{
	"R0", // REGZERO, never allocated, but keeps register numbering straight
	"SP", // REGSP
	"SB", // R2 is the thread pointer; the slot carries the SB pseudo-register
	"R3",
	"R4",
	"R5",
	"R6",
	"R7",
	"R8",
	"R9",
	"R10",
	"R11", // REGCTXT for closures
	"R12",
	"R13", // reserved (SysV small data)
	"R14",
	"R15",
	"R16",
	"R17",
	"R18",
	"R19",
	"R20",
	"R21",
	"R22",
	"R23",
	"R24",
	"R25",
	"R26",
	"R27",
	"R28",
	"R29",
	"g",   // REGG
	"R31", // REGTMP

	"F0", // FP registers are unused while floating point is soft
}

func init() {
	// Make map from reg names to reg integers.
	if len(regNamesPPC) > 64 {
		panic("too many registers")
	}
	num := map[string]int{}
	for i, name := range regNamesPPC {
		num[name] = i
	}
	buildReg := func(s string) regMask {
		m := regMask{}
		for _, r := range strings.Split(s, " ") {
			if n, ok := num[r]; ok {
				m = m.addReg(uint(n))
				continue
			}
			panic("register " + r + " not found")
		}
		return m
	}

	// Common individual register masks
	var (
		gp         = buildReg("R3 R4 R5 R6 R7 R8 R9 R10 R11 R12 R14 R15 R16 R17 R18 R19 R20 R21 R22 R23 R24 R25 R26 R27 R28 R29")
		gpg        = gp.union(buildReg("g"))
		gpsp       = gp.union(buildReg("SP"))
		gpspg      = gpg.union(buildReg("SP"))
		gpspsbg    = gpspg.union(buildReg("SB"))
		fp         = regMask{}
		ctxt       = buildReg("R11")
		callerSave = gp.union(buildReg("g")) // runtime.setg (and anything calling it) may clobber g
		first7     = buildReg("R3 R4 R5 R6 R7 R8 R9")
		first4     = buildReg("R3 R4 R5 R6")
	)
	// Common regInfo
	var (
		gp01     = regInfo{inputs: nil, outputs: []regMask{gp}}
		gp11     = regInfo{inputs: []regMask{gpg}, outputs: []regMask{gp}}
		gp11sp   = regInfo{inputs: []regMask{gpspg}, outputs: []regMask{gp}}
		gp21     = regInfo{inputs: []regMask{gpg, gpg}, outputs: []regMask{gp}}
		gp31     = regInfo{inputs: []regMask{gp, gp, gp}, outputs: []regMask{gp}}
		gp22     = regInfo{inputs: []regMask{gpg, gpg}, outputs: []regMask{gp, gp}}
		gpload   = regInfo{inputs: []regMask{gpspsbg}, outputs: []regMask{gp}}
		gpstore  = regInfo{inputs: []regMask{gpspsbg, gpg}}
		gpstore0 = regInfo{inputs: []regMask{gpspsbg}}
	)
	ops := []opData{
		{name: "ADD", argLength: 2, reg: gp21, asm: "ADD", commutative: true, earlyOk: true},                           // arg0 + arg1
		{name: "ADDconst", argLength: 1, reg: gp11sp, asm: "ADD", aux: "Int32", earlyOk: true},                         // arg0 + auxInt
		{name: "SUB", argLength: 2, reg: gp21, asm: "SUB", earlyOk: true},                                              // arg0 - arg1
		{name: "SUBconst", argLength: 1, reg: gp11, asm: "SUB", aux: "Int32", earlyOk: true},                           // arg0 - auxInt
		{name: "MUL", argLength: 2, reg: gp21, asm: "MULLW", commutative: true, earlyOk: true},                            // arg0 * arg1
		{name: "MULT", argLength: 2, reg: gp22, resultNotInArgs: true, commutative: true, typ: "(Int32,Int32)", earlyOk: true},     // arg0 * arg1, signed, results high,low
		{name: "MULTU", argLength: 2, reg: gp22, resultNotInArgs: true, commutative: true, typ: "(UInt32,UInt32)", earlyOk: true}, // arg0 * arg1, unsigned, results high,low
		{name: "DIV", argLength: 2, reg: gp22, resultNotInArgs: true, typ: "(Int32,Int32)"},                                        // arg0 / arg1, signed, results arg0%arg1,arg0/arg1
		{name: "DIVU", argLength: 2, reg: gp22, resultNotInArgs: true, typ: "(UInt32,UInt32)"},                                    // arg0 / arg1, signed, results arg0%arg1,arg0/arg1


		{name: "AND", argLength: 2, reg: gp21, asm: "AND", commutative: true, earlyOk: true},                // arg0 & arg1
		{name: "ANDconst", argLength: 1, reg: gp11, asm: "AND", aux: "Int32", earlyOk: true},                // arg0 & auxInt
		{name: "OR", argLength: 2, reg: gp21, asm: "OR", commutative: true, earlyOk: true},                  // arg0 | arg1
		{name: "ORconst", argLength: 1, reg: gp11, asm: "OR", aux: "Int32", earlyOk: true},                  // arg0 | auxInt
		{name: "XOR", argLength: 2, reg: gp21, asm: "XOR", commutative: true, typ: "UInt32", earlyOk: true}, // arg0 ^ arg1
		{name: "XORconst", argLength: 1, reg: gp11, asm: "XOR", aux: "Int32", typ: "UInt32", earlyOk: true}, // arg0 ^ auxInt
		{name: "NOR", argLength: 2, reg: gp21, asm: "NOR", commutative: true, earlyOk: true},                // ^(arg0 | arg1)

		{name: "NEG", argLength: 1, reg: gp11, earlyOk: true},                 // -arg0

		// shifts
		{name: "SLL", argLength: 2, reg: gp21, asm: "SLW", earlyOk: true},                    // arg0 << arg1, shift amount is mod 32
		{name: "SLLconst", argLength: 1, reg: gp11, asm: "SLW", aux: "Int32", earlyOk: true}, // arg0 << auxInt, shift amount must be 0 through 31 inclusive
		{name: "SRL", argLength: 2, reg: gp21, asm: "SRW", earlyOk: true},                    // arg0 >> arg1, unsigned, shift amount is mod 32
		{name: "SRLconst", argLength: 1, reg: gp11, asm: "SRW", aux: "Int32", earlyOk: true}, // arg0 >> auxInt, shift amount must be 0 through 31 inclusive
		{name: "SRA", argLength: 2, reg: gp21, asm: "SRAW", earlyOk: true},                    // arg0 >> arg1, signed, shift amount is mod 32
		{name: "SRAconst", argLength: 1, reg: gp11, asm: "SRAW", aux: "Int32", earlyOk: true}, // arg0 >> auxInt, signed, shift amount must be 0 through 31 inclusive

		{name: "CLZ", argLength: 1, reg: gp11, asm: "CNTLZW", earlyOk: true},

		// comparisons
		{name: "SGT", argLength: 2, reg: gp21, typ: "Bool", earlyOk: true},                      // 1 if arg0 > arg1 (signed), 0 otherwise
		{name: "SGTconst", argLength: 1, reg: gp11, aux: "Int32", typ: "Bool", earlyOk: true},   // 1 if auxInt > arg0 (signed), 0 otherwise
		{name: "SGTzero", argLength: 1, reg: gp11, typ: "Bool", earlyOk: true},                  // 1 if arg0 > 0 (signed), 0 otherwise
		{name: "SGTU", argLength: 2, reg: gp21, typ: "Bool", earlyOk: true},                    // 1 if arg0 > arg1 (unsigned), 0 otherwise
		{name: "SGTUconst", argLength: 1, reg: gp11, aux: "Int32", typ: "Bool", earlyOk: true}, // 1 if auxInt > arg0 (unsigned), 0 otherwise
		{name: "SGTUzero", argLength: 1, reg: gp11, typ: "Bool", earlyOk: true},                // 1 if arg0 > 0 (unsigned), 0 otherwise


		// moves
		{name: "MOVWconst", argLength: 0, reg: gp01, aux: "Int32", asm: "MOVW", typ: "UInt32", rematerializeable: true, earlyOk: true},    // auxint

		{name: "MOVWaddr", argLength: 1, reg: regInfo{inputs: []regMask{buildReg("SP").union(buildReg("SB"))}, outputs: []regMask{gp}}, aux: "SymOff", asm: "MOVW", rematerializeable: true, symEffect: "Addr", earlyOk: true}, // arg0 + auxInt + aux.(*gc.Sym), arg0=SP/SB

		{name: "MOVBload", argLength: 2, reg: gpload, aux: "SymOff", asm: "MOVB", typ: "Int8", faultOnNilArg0: true, symEffect: "Read", addrSinkArg0: true},     // load from arg0 + auxInt + aux.  arg1=mem.
		{name: "MOVBUload", argLength: 2, reg: gpload, aux: "SymOff", asm: "MOVBZ", typ: "UInt8", faultOnNilArg0: true, symEffect: "Read", addrSinkArg0: true},  // load from arg0 + auxInt + aux.  arg1=mem.
		{name: "MOVHload", argLength: 2, reg: gpload, aux: "SymOff", asm: "MOVH", typ: "Int16", faultOnNilArg0: true, symEffect: "Read", addrSinkArg0: true},    // load from arg0 + auxInt + aux.  arg1=mem.
		{name: "MOVHUload", argLength: 2, reg: gpload, aux: "SymOff", asm: "MOVHZ", typ: "UInt16", faultOnNilArg0: true, symEffect: "Read", addrSinkArg0: true}, // load from arg0 + auxInt + aux.  arg1=mem.
		{name: "MOVWload", argLength: 2, reg: gpload, aux: "SymOff", asm: "MOVWZ", typ: "UInt32", faultOnNilArg0: true, symEffect: "Read", addrSinkArg0: true},   // load from arg0 + auxInt + aux.  arg1=mem.

		{name: "MOVBstore", argLength: 3, reg: gpstore, aux: "SymOff", asm: "MOVB", typ: "Mem", faultOnNilArg0: true, symEffect: "Write", addrSinkArg0: true}, // store 1 byte of arg1 to arg0 + auxInt + aux.  arg2=mem.
		{name: "MOVHstore", argLength: 3, reg: gpstore, aux: "SymOff", asm: "MOVH", typ: "Mem", faultOnNilArg0: true, symEffect: "Write", addrSinkArg0: true}, // store 2 bytes of arg1 to arg0 + auxInt + aux.  arg2=mem.
		{name: "MOVWstore", argLength: 3, reg: gpstore, aux: "SymOff", asm: "MOVW", typ: "Mem", faultOnNilArg0: true, symEffect: "Write", addrSinkArg0: true}, // store 4 bytes of arg1 to arg0 + auxInt + aux.  arg2=mem.

		{name: "MOVBstorezero", argLength: 2, reg: gpstore0, aux: "SymOff", asm: "MOVB", typ: "Mem", faultOnNilArg0: true, symEffect: "Write", addrSinkArg0: true}, // store 1 byte of zero to arg0 + auxInt + aux.  arg1=mem.
		{name: "MOVHstorezero", argLength: 2, reg: gpstore0, aux: "SymOff", asm: "MOVH", typ: "Mem", faultOnNilArg0: true, symEffect: "Write", addrSinkArg0: true}, // store 2 bytes of zero to arg0 + auxInt + aux.  arg1=mem.
		{name: "MOVWstorezero", argLength: 2, reg: gpstore0, aux: "SymOff", asm: "MOVW", typ: "Mem", faultOnNilArg0: true, symEffect: "Write", addrSinkArg0: true}, // store 4 bytes of zero to arg0 + auxInt + aux.  arg1=mem.

		// moves (no conversion)

		// conversions
		{name: "MOVBreg", argLength: 1, reg: gp11, asm: "MOVB", earlyOk: true},   // move from arg0, sign-extended from byte
		{name: "MOVBUreg", argLength: 1, reg: gp11, asm: "MOVBZ", earlyOk: true}, // move from arg0, unsign-extended from byte
		{name: "MOVHreg", argLength: 1, reg: gp11, asm: "MOVH", earlyOk: true},   // move from arg0, sign-extended from half
		{name: "MOVHUreg", argLength: 1, reg: gp11, asm: "MOVHZ", earlyOk: true}, // move from arg0, unsign-extended from half
		{name: "MOVWreg", argLength: 1, reg: gp11, asm: "MOVW", earlyOk: true},   // move from arg0

		{name: "MOVWnop", argLength: 1, reg: regInfo{inputs: []regMask{gp}, outputs: []regMask{gp}}, resultInArg0: true, earlyOk: true}, // nop, return arg0 in same register

		// conditional move on zero (returns arg1 if arg2 is 0, otherwise arg0)
		// order of parameters is reversed so we can use resultInArg0 (OpCMOVZ result arg1 arg2-> CMOVZ arg2reg, arg1reg, resultReg)
		{name: "CMOVZ", argLength: 3, reg: gp31, resultInArg0: true, earlyOk: true},
		{name: "CMOVZzero", argLength: 2, reg: regInfo{inputs: []regMask{gp, gpg}, outputs: []regMask{gp}}, resultInArg0: true, earlyOk: true},


		// function calls
		{name: "CALLstatic", argLength: 1, reg: regInfo{clobbers: callerSave}, aux: "CallOff", clobberFlags: true, call: true},                                                       // call static function aux.(*obj.LSym).  arg0=mem, auxint=argsize, returns mem
		{name: "CALLtail", argLength: 1, reg: regInfo{clobbers: callerSave}, aux: "CallOff", clobberFlags: true, call: true, tailCall: true},                                         //  tail call static function aux.(*obj.LSym).  arg0=mem, auxint=argsize, returns mem
		{name: "CALLtailinter", argLength: 2, reg: regInfo{inputs: []regMask{gp}, clobbers: callerSave}, aux: "CallOff", clobberFlags: true, call: true, tailCall: true},             //  tail call fn by pointer.  arg0=cpdeptr, arg1=mem, auxint=argsize, returns mem
		{name: "CALLclosure", argLength: 3, reg: regInfo{inputs: []regMask{gpsp, ctxt, regMask{}}, clobbers: callerSave}, aux: "CallOff", clobberFlags: true, call: true}, // call function via closure.  arg0=codeptr, arg1=closure, arg2=mem, auxint=argsize, returns mem
		{name: "CALLinter", argLength: 2, reg: regInfo{inputs: []regMask{gp}, clobbers: callerSave}, aux: "CallOff", clobberFlags: true, call: true},                                 // call fn by pointer.  arg0=codeptr, arg1=mem, auxint=argsize, returns mem

		// atomic ops

		// load from arg0. arg1=mem.
		// returns <value,memory> so they can be properly ordered with other loads.
		// SYNC
		// MOV(B|W)	(Rarg0), Rout
		// SYNC

		// store arg1 to arg0. arg2=mem. returns memory.
		// SYNC
		// MOV(B|W)	Rarg1, (Rarg0)
		// SYNC

		// atomic exchange.
		// store arg1 to arg0. arg2=mem. returns <old content of *arg0, memory>.
		// SYNC
		// LL	(Rarg0), Rout
		// MOVW Rarg1, Rtmp
		// SC	Rtmp, (Rarg0)
		// BEQ	Rtmp, -3(PC)
		// SYNC

		// atomic add.
		// *arg0 += arg1. arg2=mem. returns <new content of *arg0, memory>.
		// SYNC
		// LL	(Rarg0), Rout
		// ADDU Rarg1, Rout, Rtmp
		// SC	Rtmp, (Rarg0)
		// BEQ	Rtmp, -3(PC)
		// SYNC
		// ADDU Rarg1, Rout

		// atomic compare and swap.
		// arg0 = pointer, arg1 = old value, arg2 = new value, arg3 = memory.
		// if *arg0 == arg1 {
		//   *arg0 = arg2
		//   return (true, memory)
		// } else {
		//   return (false, memory)
		// }
		// SYNC
		// MOVW $0, Rout
		// LL	(Rarg0), Rtmp
		// BNE	Rtmp, Rarg1, 4(PC)
		// MOVW Rarg2, Rout
		// SC	Rout, (Rarg0)
		// BEQ	Rout, -4(PC)
		// SYNC

		// atomic and/or.
		// *arg0 &= (|=) arg1. arg2=mem. returns memory.
		// SYNC
		// LL	(Rarg0), Rtmp
		// AND	Rarg1, Rtmp
		// SC	Rtmp, (Rarg0)
		// BEQ	Rtmp, -3(PC)
		// SYNC

		// large or unaligned zeroing
		// arg0 = address of memory to zero (in R1, changed as side effect)
		// arg1 = address of the last element to zero
		// arg2 = mem
		// auxint = alignment
		// returns mem
		//	SUBU	$4, R1
		//	MOVW	R0, 4(R1)
		//	ADDU	$4, R1
		//	BNE	Rarg1, R1, -2(PC)
		{
			name:      "LoweredZero",
			aux:       "Int32",
			argLength: 3,
			reg: regInfo{
				inputs:   []regMask{buildReg("R3"), gp},
				clobbers: buildReg("R3"),
			},
			faultOnNilArg0: true,
			addrSinkArg0:   true,
			addrSinkArg1:   true,
		},

		// large or unaligned move
		// arg0 = address of dst memory (in R2, changed as side effect)
		// arg1 = address of src memory (in R1, changed as side effect)
		// arg2 = address of the last element of src
		// arg3 = mem
		// auxint = alignment
		// returns mem
		//	SUBU	$4, R1
		//	MOVW	4(R1), Rtmp
		//	MOVW	Rtmp, (R2)
		//	ADDU	$4, R1
		//	ADDU	$4, R2
		//	BNE	Rarg2, R1, -4(PC)
		{
			name:      "LoweredMove",
			aux:       "Int32",
			argLength: 4,
			reg: regInfo{
				inputs:   []regMask{buildReg("R3"), buildReg("R4"), gp},
				clobbers: buildReg("R3 R4"),
			},
			faultOnNilArg0: true,
			faultOnNilArg1: true,
			addrSinkArg0:   true,
			addrSinkArg1:   true,
			// TODO: could use addrSinkArg2 here.
		},

		// pseudo-ops
		{name: "LoweredNilCheck", argLength: 2, reg: regInfo{inputs: []regMask{gpg}}, nilCheck: true, faultOnNilArg0: true}, // panic if arg0 is nil.  arg1=mem.


		// Scheduler ensures LoweredGetClosurePtr occurs only in entry block,
		// and sorts it to the very beginning of the block to prevent other
		// use of R11 (ppc64.REGCTXT, the closure pointer)
		{name: "LoweredGetClosurePtr", reg: regInfo{outputs: []regMask{ctxt}}, zeroWidth: true},

		// LoweredGetCallerSP returns the SP of the caller of the current function. arg0=mem.
		{name: "LoweredGetCallerSP", argLength: 1, reg: gp01, rematerializeable: true},

		// LoweredGetCallerPC evaluates to the PC to which its "caller" will return.
		// I.e., if f calls g "calls" sys.GetCallerPC,
		// the result should be the PC within f that g will return to.
		// See runtime/stubs.go for a more detailed discussion.
		{name: "LoweredGetCallerPC", reg: gp01, rematerializeable: true},

		// LoweredWB invokes runtime.gcWriteBarrier. arg0=mem, auxint=# of buffer entries needed
		// Like ppc64, it preserves R3-R10, R14-R17, R20, R21 and g and
		// clobbers the rest of the allocatable registers, LR and R31 (REGTMP).
		// Returns a pointer to a write barrier buffer in R29.
		{name: "LoweredWB", argLength: 1, reg: regInfo{clobbers: callerSave.minus(buildReg("R3 R4 R5 R6 R7 R8 R9 R10 R14 R15 R16 R17 R20 R21 g")), outputs: []regMask{buildReg("R29")}}, clobberFlags: true, aux: "Int64"},

		// Do data barrier. arg0=memorys
		{name: "LoweredPubBarrier", argLength: 1, asm: "SYNC", hasSideEffects: true},

		// LoweredPanicBoundsRR takes x and y, two values that caused a bounds check to fail.
		// the RC and CR versions are used when one of the arguments is a constant. CC is used
		// when both are constant (normally both 0, as prove derives the fact that a [0] bounds
		// failure means the length must have also been 0).
		// AuxInt contains a report code (see PanicBounds in genericOps.go).
		{name: "LoweredPanicBoundsRR", argLength: 3, aux: "Int64", reg: regInfo{inputs: []regMask{first7, first7}}, typ: "Mem", call: true}, // arg0=x, arg1=y, arg2=mem, returns memory.
		{name: "LoweredPanicBoundsRC", argLength: 2, aux: "PanicBoundsC", reg: regInfo{inputs: []regMask{first7}}, typ: "Mem", call: true},   // arg0=x, arg1=mem, returns memory.
		{name: "LoweredPanicBoundsCR", argLength: 2, aux: "PanicBoundsC", reg: regInfo{inputs: []regMask{first7}}, typ: "Mem", call: true},   // arg0=y, arg1=mem, returns memory.
		{name: "LoweredPanicBoundsCC", argLength: 1, aux: "PanicBoundsCC", reg: regInfo{}, typ: "Mem", call: true},                            // arg0=mem, returns memory.

		// Same as above, but the x value is 64 bits.
		{name: "LoweredPanicExtendRR", argLength: 4, aux: "Int64", reg: regInfo{inputs: []regMask{first4, first4, first7}}, typ: "Mem", call: true}, // arg0=x_hi, arg1=x_lo, arg2=y, arg3=mem, returns memory.
		{name: "LoweredPanicExtendRC", argLength: 3, aux: "PanicBoundsC", reg: regInfo{inputs: []regMask{first4, first4}}, typ: "Mem", call: true},   // arg0=x_hi, arg1=x_lo, arg2=mem, returns memory.
	}

	blocks := []blockData{
		{name: "EQ", controls: 1},
		{name: "NE", controls: 1},
		{name: "LTZ", controls: 1}, // < 0
		{name: "LEZ", controls: 1}, // <= 0
		{name: "GTZ", controls: 1}, // > 0
		{name: "GEZ", controls: 1}, // >= 0

		// JUMPTABLE implements jump tables.
		// Aux is the symbol (an *obj.LSym) for the jump table.
		// control[0] is the index into the jump table.
		// control[1] is the address of the jump table (the address of the symbol stored in Aux).
		{name: "JUMPTABLE", controls: 2, aux: "Sym"},
	}

	archs = append(archs, arch{
		name:            "PPC",
		pkg:             "cmd/internal/obj/ppc64",
		genfile:         "../../ppc/ssa.go",
		ops:             ops,
		blocks:          blocks,
		regnames:        regNamesPPC,
		gpregmask:       gp,
		fpregmask:       fp,
		framepointerreg: -1, // not used
		linkreg:         -1, // not used
	})
}
