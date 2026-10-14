// Code generated from _gen/PPClatelower.rules using 'go generate'; DO NOT EDIT.

package ppclatelower

import "cmd/compile/internal/types"
import "cmd/compile/internal/ssa/block"
import "cmd/compile/internal/ssa/ssaop"
import "cmd/compile/internal/ssa"

func RewriteValue(v *ssa.Value) bool {
	return false
}
func RewriteBlock(b *ssa.Block) bool {
	typ := &b.Func.Config.Types
	switch b.Kind {
	case block.BlockPPCEQ:
		// match: (EQ (SGT x y) yes no)
		// result: (CLE (CMPW x y) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGT {
			v_0 := b.Controls[0]
			y := v_0.Args[1]
			x := v_0.Args[0]
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPW, types.TypeFlags)
			v0.AddArg2(x, y)
			b.ResetWithControl(block.BlockPPCCLE, v0)
			return true
		}
		// match: (EQ (SGTU x y) yes no)
		// result: (CLE (CMPWU x y) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTU {
			v_0 := b.Controls[0]
			y := v_0.Args[1]
			x := v_0.Args[0]
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPWU, types.TypeFlags)
			v0.AddArg2(x, y)
			b.ResetWithControl(block.BlockPPCCLE, v0)
			return true
		}
		// match: (EQ (SGTconst [c] x) yes no)
		// cond: ssa.Is16Bit(int64(c))
		// result: (CGE (CMPWconst [c] x) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTconst {
			v_0 := b.Controls[0]
			c := ssa.AuxIntToInt32(v_0.AuxInt)
			x := v_0.Args[0]
			if !(ssa.Is16Bit(int64(c))) {
				break
			}
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(c)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCGE, v0)
			return true
		}
		// match: (EQ (SGTUconst [c] x) yes no)
		// cond: uint32(c) <= 0xffff
		// result: (CGE (CMPWUconst [c] x) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTUconst {
			v_0 := b.Controls[0]
			c := ssa.AuxIntToInt32(v_0.AuxInt)
			x := v_0.Args[0]
			if !(uint32(c) <= 0xffff) {
				break
			}
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPWUconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(c)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCGE, v0)
			return true
		}
		// match: (EQ (SGTconst [c] x) yes no)
		// result: (CGE (CMPW x (MOVWconst [c])) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTconst {
			v_0 := b.Controls[0]
			c := ssa.AuxIntToInt32(v_0.AuxInt)
			x := v_0.Args[0]
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPW, types.TypeFlags)
			v1 := b.NewValue0(v_0.Pos, ssaop.OpPPCMOVWconst, typ.UInt32)
			v1.AuxInt = ssa.Int32ToAuxInt(c)
			v0.AddArg2(x, v1)
			b.ResetWithControl(block.BlockPPCCGE, v0)
			return true
		}
		// match: (EQ (SGTUconst [c] x) yes no)
		// result: (CGE (CMPWU x (MOVWconst [c])) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTUconst {
			v_0 := b.Controls[0]
			c := ssa.AuxIntToInt32(v_0.AuxInt)
			x := v_0.Args[0]
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPWU, types.TypeFlags)
			v1 := b.NewValue0(v_0.Pos, ssaop.OpPPCMOVWconst, typ.UInt32)
			v1.AuxInt = ssa.Int32ToAuxInt(c)
			v0.AddArg2(x, v1)
			b.ResetWithControl(block.BlockPPCCGE, v0)
			return true
		}
		// match: (EQ (SGTzero x) yes no)
		// result: (CLE (CMPWconst [0] x) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTzero {
			v_0 := b.Controls[0]
			x := v_0.Args[0]
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(0)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCLE, v0)
			return true
		}
		// match: (EQ (SGTUzero x) yes no)
		// result: (CEQ (CMPWconst [0] x) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTUzero {
			v_0 := b.Controls[0]
			x := v_0.Args[0]
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(0)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCEQ, v0)
			return true
		}
		// match: (EQ (FEqual cmp) yes no)
		// result: (CNE cmp yes no)
		for b.Controls[0].Op == ssaop.OpPPCFEqual {
			v_0 := b.Controls[0]
			cmp := v_0.Args[0]
			b.ResetWithControl(block.BlockPPCCNE, cmp)
			return true
		}
		// match: (EQ x yes no)
		// result: (CEQ (CMPWconst [0] x) yes no)
		for {
			x := b.Controls[0]
			v0 := b.NewValue0(x.Pos, ssaop.OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(0)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCEQ, v0)
			return true
		}
	case block.BlockPPCGEZ:
		// match: (GEZ x yes no)
		// result: (CGE (CMPWconst [0] x) yes no)
		for {
			x := b.Controls[0]
			v0 := b.NewValue0(x.Pos, ssaop.OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(0)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCGE, v0)
			return true
		}
	case block.BlockPPCGTZ:
		// match: (GTZ x yes no)
		// result: (CGT (CMPWconst [0] x) yes no)
		for {
			x := b.Controls[0]
			v0 := b.NewValue0(x.Pos, ssaop.OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(0)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCGT, v0)
			return true
		}
	case block.BlockPPCLEZ:
		// match: (LEZ x yes no)
		// result: (CLE (CMPWconst [0] x) yes no)
		for {
			x := b.Controls[0]
			v0 := b.NewValue0(x.Pos, ssaop.OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(0)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCLE, v0)
			return true
		}
	case block.BlockPPCLTZ:
		// match: (LTZ x yes no)
		// result: (CLT (CMPWconst [0] x) yes no)
		for {
			x := b.Controls[0]
			v0 := b.NewValue0(x.Pos, ssaop.OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(0)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCLT, v0)
			return true
		}
	case block.BlockPPCNE:
		// match: (NE (SGT x y) yes no)
		// result: (CGT (CMPW x y) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGT {
			v_0 := b.Controls[0]
			y := v_0.Args[1]
			x := v_0.Args[0]
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPW, types.TypeFlags)
			v0.AddArg2(x, y)
			b.ResetWithControl(block.BlockPPCCGT, v0)
			return true
		}
		// match: (NE (SGTU x y) yes no)
		// result: (CGT (CMPWU x y) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTU {
			v_0 := b.Controls[0]
			y := v_0.Args[1]
			x := v_0.Args[0]
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPWU, types.TypeFlags)
			v0.AddArg2(x, y)
			b.ResetWithControl(block.BlockPPCCGT, v0)
			return true
		}
		// match: (NE (SGTconst [c] x) yes no)
		// cond: ssa.Is16Bit(int64(c))
		// result: (CLT (CMPWconst [c] x) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTconst {
			v_0 := b.Controls[0]
			c := ssa.AuxIntToInt32(v_0.AuxInt)
			x := v_0.Args[0]
			if !(ssa.Is16Bit(int64(c))) {
				break
			}
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(c)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCLT, v0)
			return true
		}
		// match: (NE (SGTUconst [c] x) yes no)
		// cond: uint32(c) <= 0xffff
		// result: (CLT (CMPWUconst [c] x) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTUconst {
			v_0 := b.Controls[0]
			c := ssa.AuxIntToInt32(v_0.AuxInt)
			x := v_0.Args[0]
			if !(uint32(c) <= 0xffff) {
				break
			}
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPWUconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(c)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCLT, v0)
			return true
		}
		// match: (NE (SGTconst [c] x) yes no)
		// result: (CLT (CMPW x (MOVWconst [c])) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTconst {
			v_0 := b.Controls[0]
			c := ssa.AuxIntToInt32(v_0.AuxInt)
			x := v_0.Args[0]
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPW, types.TypeFlags)
			v1 := b.NewValue0(v_0.Pos, ssaop.OpPPCMOVWconst, typ.UInt32)
			v1.AuxInt = ssa.Int32ToAuxInt(c)
			v0.AddArg2(x, v1)
			b.ResetWithControl(block.BlockPPCCLT, v0)
			return true
		}
		// match: (NE (SGTUconst [c] x) yes no)
		// result: (CLT (CMPWU x (MOVWconst [c])) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTUconst {
			v_0 := b.Controls[0]
			c := ssa.AuxIntToInt32(v_0.AuxInt)
			x := v_0.Args[0]
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPWU, types.TypeFlags)
			v1 := b.NewValue0(v_0.Pos, ssaop.OpPPCMOVWconst, typ.UInt32)
			v1.AuxInt = ssa.Int32ToAuxInt(c)
			v0.AddArg2(x, v1)
			b.ResetWithControl(block.BlockPPCCLT, v0)
			return true
		}
		// match: (NE (SGTzero x) yes no)
		// result: (CGT (CMPWconst [0] x) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTzero {
			v_0 := b.Controls[0]
			x := v_0.Args[0]
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(0)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCGT, v0)
			return true
		}
		// match: (NE (SGTUzero x) yes no)
		// result: (CNE (CMPWconst [0] x) yes no)
		for b.Controls[0].Op == ssaop.OpPPCSGTUzero {
			v_0 := b.Controls[0]
			x := v_0.Args[0]
			v0 := b.NewValue0(v_0.Pos, ssaop.OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(0)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCNE, v0)
			return true
		}
		// match: (NE (FEqual cmp) yes no)
		// result: (CEQ cmp yes no)
		for b.Controls[0].Op == ssaop.OpPPCFEqual {
			v_0 := b.Controls[0]
			cmp := v_0.Args[0]
			b.ResetWithControl(block.BlockPPCCEQ, cmp)
			return true
		}
		// match: (NE (FLessThan cmp) yes no)
		// result: (CLT cmp yes no)
		for b.Controls[0].Op == ssaop.OpPPCFLessThan {
			v_0 := b.Controls[0]
			cmp := v_0.Args[0]
			b.ResetWithControl(block.BlockPPCCLT, cmp)
			return true
		}
		// match: (NE (FGreaterThan cmp) yes no)
		// result: (CGT cmp yes no)
		for b.Controls[0].Op == ssaop.OpPPCFGreaterThan {
			v_0 := b.Controls[0]
			cmp := v_0.Args[0]
			b.ResetWithControl(block.BlockPPCCGT, cmp)
			return true
		}
		// match: (NE x yes no)
		// result: (CNE (CMPWconst [0] x) yes no)
		for {
			x := b.Controls[0]
			v0 := b.NewValue0(x.Pos, ssaop.OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = ssa.Int32ToAuxInt(0)
			v0.AddArg(x)
			b.ResetWithControl(block.BlockPPCCNE, v0)
			return true
		}
	}
	return false
}
