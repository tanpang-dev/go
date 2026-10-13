// 64bit 整数演算の差分テスト: arm64 と ppc で同じ出力になるはず
package main

import (
	"fmt"
	"time"
)

var vals = []int64{0, 1, -1, 2, -2, 7, -7, 1000, -1000, 50000000, 8221920, 1 << 31, -(1 << 31), 1<<32 - 1, 1 << 32, 1<<40 + 12345, -(1<<40 + 12345), 1<<62 + 3, -(1 << 62), 9223372036854775807, -9223372036854775808}

//go:noinline
func ops(a, b int64) string {
	ua, ub := uint64(a), uint64(b)
	s := fmt.Sprint(a+b, a-b, a*b, a<b, a<=b, a>b, a>=b, a==b, ua<ub, ua>ub, a&b, a|b, a^b, a<<3, a>>3, ua>>3)
	if b != 0 && !(a == -9223372036854775808 && b == -1) {
		s += fmt.Sprint(" ", a/b, a%b, ua/ub, ua%ub)
	}
	for _, sh := range []uint{0, 1, 31, 32, 33, 63, 64, 100} {
		s += fmt.Sprint(" ", a<<sh, a>>sh, ua>>sh)
	}
	return s
}

func main() {
	h := uint64(14695981039346656037)
	for _, a := range vals {
		for _, b := range vals {
			s := ops(a, b)
			for i := 0; i < len(s); i++ {
				h ^= uint64(s[i])
				h *= 1099511628211
			}
		}
	}
	fmt.Printf("hash=%x\n", h)
	for _, a := range vals[:12] {
		fmt.Println(a, ops(a, 7))
	}
	d := 50*time.Millisecond + 8221920
	fmt.Println(d, d.Round(time.Millisecond), d >= 50*time.Millisecond, d.Nanoseconds())
}
