// 浮動小数点の差分テスト: arm64 と ppc でビット単位まで同じ出力になるはず
package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

var f64s = []float64{0, math.Copysign(0, -1), 1, -1, 0.1, -2.5, 3.0000001, 1e-300, 1e300, 123456789.123, -987654.321, math.Pi, math.E, math.MaxFloat64, math.SmallestNonzeroFloat64, math.Inf(1), math.Inf(-1), math.NaN(), 2147483647.9, -2147483648.5, 4294967295.0, 1 << 53, 0.5, 1.5, 2.5}
var i32s = []int32{0, 1, -1, 7, -7, 1 << 30, math.MaxInt32, math.MinInt32, 123456789, -16777217}
var u32s = []uint32{0, 1, 1 << 31, math.MaxUint32, 16777217, 3000000000}
var i64s = []int64{0, 1, -1, 1 << 53, 1<<53 + 1, math.MaxInt64, math.MinInt64, -123456789012345}

func bits(x float64) string { return strconv.FormatUint(math.Float64bits(x), 16) }
func bits32(x float32) string { return strconv.FormatUint(uint64(math.Float32bits(x)), 16) }

//go:noinline
func ops(a, b float64) string {
	fa, fb := float32(a), float32(b)
	return fmt.Sprint(bits(a+b), bits(a-b), bits(a*b), bits(a/b), a < b, a <= b, a > b, a >= b, a == b, a != b,
		bits32(fa+fb), bits32(fa-fb), bits32(fa*fb), bits32(fa/fb), fa < fb, fa == fb, bits(-a), bits(math.Abs(a)))
}

var verbose = len(os.Args) > 1

func inRange32(x float64) bool { return x > -2147483649 && x < 2147483648 }

func main() {
	h := uint64(14695981039346656037)
	add := func(s string) {
		if verbose {
			fmt.Println(s)
		}
		for i := 0; i < len(s); i++ {
			h ^= uint64(s[i])
			h *= 1099511628211
		}
	}
	for _, a := range f64s {
		for _, b := range f64s {
			add(ops(a, b))
		}
	}
	for _, x := range f64s {
		if inRange32(x) {
			add(fmt.Sprint(int32(x), int32(float32(x))))
		}
		if x >= 0 && x < 4294967296 {
			add(fmt.Sprint(uint32(x)))
		}
		if x > -9.2e18 && x < 9.2e18 {
			add(fmt.Sprint(int64(x)))
		}
		add(strconv.FormatFloat(x, 'g', -1, 64) + strconv.FormatFloat(x, 'e', 17, 64) + strconv.FormatFloat(float64(float32(x)), 'g', -1, 32))
		add(fmt.Sprint(bits(math.Sqrt(x)), bits(math.Floor(x)), bits(math.Ceil(x)), bits(math.Trunc(x))))
		if !math.IsInf(x, 0) && math.Abs(x) < 1e6 {
			add(fmt.Sprint(bits(math.Sin(x)), bits(math.Exp(x/1000)), bits(math.Log(math.Abs(x)+1)), bits(math.Pow(math.Abs(x)+1, 0.37))))
		}
	}
	for _, i := range i32s {
		add(fmt.Sprint(bits(float64(i)), bits32(float32(i))))
	}
	for _, u := range u32s {
		add(fmt.Sprint(bits(float64(u)), bits32(float32(u))))
	}
	for _, i := range i64s {
		add(fmt.Sprint(bits(float64(i)), bits32(float32(i)), bits(float64(uint64(i)))))
	}
	for _, s := range []string{"3.14159", "1e-7", "6.02214076e23", "-0", "NaN", "Inf", "0x1p-1074", "123456789012345678901234567890"} {
		v, err := strconv.ParseFloat(s, 64)
		w, _ := strconv.ParseFloat(s, 32)
		add(fmt.Sprint(bits(v), bits(w), err))
	}
	fmt.Printf("hash=%x\n", h)
	fmt.Println(math.Sqrt(2), math.Sin(1), float32(1)/3, int32(f64s[11]*1e6), uint32(f64s[22]*1e9), float64(uint32(4294967295)), int64(f64s[9]*1e3))
}
