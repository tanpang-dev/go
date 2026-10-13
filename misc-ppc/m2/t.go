// M2 確認用: runtime を使わない範囲の Go コードを ppc でコンパイルし、結果を out に書く。
// 偽 runtime(rt.s) が main.main を呼んでから out を標準出力に書く。
package main

var out [256]byte
var n int

func put(s string) {
	for i := 0; i < len(s); i++ {
		out[n] = s[i]
		n++
	}
}

func putu(x uint32) {
	var b [10]byte
	i := len(b)
	for {
		i--
		b[i] = byte('0' + x%10)
		x /= 10
		if x == 0 {
			break
		}
	}
	for ; i < len(b); i++ {
		out[n] = b[i]
		n++
	}
}

func puti(x int32) {
	if x < 0 {
		put("-")
		putu(uint32(-x))
		return
	}
	putu(uint32(x))
}

//go:noinline
func fib(k int32) int32 {
	if k < 2 {
		return k
	}
	return fib(k-1) + fib(k-2)
}

//go:noinline
func mul64(a, b uint64) uint64 { return a*b + a<<7 - b>>3 }

//go:noinline
func div(a, b int32) (int32, int32) { return a / b, a % b }

//go:noinline
func shl(x uint32, s uint) uint32 { return x << s }

//go:noinline
func sar(x int32, s uint) int32 { return x >> s }

type pt struct{ x, y int32 }

//go:noinline
func sum(ps []pt) (s int32) {
	for _, p := range ps {
		s += p.x * p.y
	}
	return
}

func main() {
	put("fib(20)=")
	puti(fib(20))
	put(" mul64=")
	m := mul64(0x123456789, 0xabcdef)
	putu(uint32(m >> 32))
	put(":")
	putu(uint32(m))
	q, r := div(-1000, 7)
	put(" div=")
	puti(q)
	put(",")
	puti(r)
	q, r = div(-2147483648, -1)
	put(" minint/-1=")
	puti(q)
	put(",")
	puti(r)
	put(" shl=")
	putu(shl(1, 31))
	put(",")
	putu(shl(1, 32))
	put(",")
	putu(shl(1, 100))
	put(" sar=")
	puti(sar(-8, 1))
	put(",")
	puti(sar(-8, 40))
	ps := [3]pt{{1, 2}, {3, 4}, {5, 6}}
	put(" sum=")
	puti(sum(ps[:]))
	put("\n")
}
