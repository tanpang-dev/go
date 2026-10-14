// math.Sqrt(ppc の frsqrte+FMA 版)を big.Float(正確に丸めた値)と突き合わせる
package main

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"os"
	"strconv"
)

func ref(x float64) float64 {
	b := new(big.Float).SetPrec(200).SetFloat64(x)
	b.Sqrt(b)
	f, _ := new(big.Float).SetPrec(53).SetMode(big.ToNearestEven).Set(b).Float64()
	return f
}

func main() {
	n := 200000
	if len(os.Args) > 1 {
		n, _ = strconv.Atoi(os.Args[1])
	}
	bad := 0
	check := func(x float64) {
		got, want := math.Sqrt(x), ref(x)
		if math.Float64bits(got) != math.Float64bits(want) {
			if bad < 10 {
				fmt.Printf("x=%v (%x) got %x want %x\n", x, math.Float64bits(x), math.Float64bits(got), math.Float64bits(want))
			}
			bad++
		}
	}
	for _, x := range []float64{math.SmallestNonzeroFloat64, 2.2250738585072014e-308, 1e-300, 0.25, 1, 2, 3, 4, 1e300, math.MaxFloat64, 1 << 52, 1<<53 - 1} {
		check(x)
		check(math.Nextafter(x, 0))
		check(math.Nextafter(x, math.Inf(1)))
	}
	r := rand.New(rand.NewSource(1))
	for i := 0; i < n; i++ {
		x := math.Float64frombits(r.Uint64() &^ (1 << 63))
		if math.IsNaN(x) || math.IsInf(x, 0) {
			continue
		}
		check(x)
		check(r.Float64() * 1000)
	}
	fmt.Println(math.Sqrt(-1), math.Sqrt(math.Inf(1)), math.Sqrt(math.Copysign(0, -1)), math.Sqrt(math.NaN()), math.Sqrt(2))
	fmt.Printf("mismatches: %d / %d\n", bad, 2*n+36)
}
