// 性能改善の基準ベンチマーク。Wii で -test.benchtime を短くして回す。
package perf

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"math"
	"sort"
	"testing"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
)

var sink int
var fsink float64

func BenchmarkMemmove4K(b *testing.B) {
	src := make([]byte, 4096)
	dst := make([]byte, 4096)
	b.SetBytes(4096)
	for i := 0; i < b.N; i++ {
		copy(dst, src)
	}
}

func BenchmarkMemclr4K(b *testing.B) {
	buf := make([]byte, 4096)
	b.SetBytes(4096)
	for i := 0; i < b.N; i++ {
		clear(buf)
	}
}

func BenchmarkEqual4K(b *testing.B) {
	x := make([]byte, 4096)
	y := make([]byte, 4096)
	b.SetBytes(4096)
	for i := 0; i < b.N; i++ {
		if bytes.Equal(x, y) {
			sink++
		}
	}
}

func BenchmarkIndexByte4K(b *testing.B) {
	x := make([]byte, 4096)
	x[4095] = 1
	b.SetBytes(4096)
	for i := 0; i < b.N; i++ {
		sink += bytes.IndexByte(x, 1)
	}
}

func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func BenchmarkFib20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink += fib(20)
	}
}

func BenchmarkSortInts1K(b *testing.B) {
	src := make([]int, 1024)
	for i := range src {
		src[i] = (i * 7919) % 1024
	}
	buf := make([]int, 1024)
	for i := 0; i < b.N; i++ {
		copy(buf, src)
		sort.Ints(buf)
	}
}

type rec struct {
	Name  string
	Age   int
	Tags  []string
	Score float64
}

func BenchmarkJSONRoundTrip(b *testing.B) {
	v := rec{"wii", 20, []string{"go", "ppc", "linux"}, 99.5}
	for i := 0; i < b.N; i++ {
		data, _ := json.Marshal(v)
		var w rec
		_ = json.Unmarshal(data, &w)
	}
}

func BenchmarkSHA256_1K(b *testing.B) {
	data := make([]byte, 1024)
	b.SetBytes(1024)
	for i := 0; i < b.N; i++ {
		sha256.Sum256(data)
	}
}

func BenchmarkChaCha20Poly1305_1K(b *testing.B) {
	aead, _ := chacha20poly1305.New(make([]byte, 32))
	nonce := make([]byte, 12)
	data := make([]byte, 1024)
	out := make([]byte, 0, 1100)
	b.SetBytes(1024)
	for i := 0; i < b.N; i++ {
		out = aead.Seal(out[:0], nonce, data, nil)
	}
}

func BenchmarkTimeNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink += time.Now().Nanosecond()
	}
}

func BenchmarkFloatMath(b *testing.B) {
	x := 1.0
	for i := 0; i < b.N; i++ {
		x = math.Sqrt(x*1.0001+0.5) * 1.5
	}
	fsink = x
}
