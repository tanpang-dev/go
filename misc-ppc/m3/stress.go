// M3 確認用: runtime の主要機能をひととおり動かす。
package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func check(name string, ok bool) {
	if ok {
		fmt.Println("OK  ", name)
		return
	}
	fmt.Println("FAIL", name)
	os.Exit(1)
}

func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func deep(n int) int {
	var pad [64]byte
	pad[n%64] = byte(n)
	if n == 0 {
		return int(pad[0])
	}
	return deep(n-1) + int(pad[n%64]&1)
}

func main() {
	start := time.Now()

	// goroutines + channels + WaitGroup
	var wg sync.WaitGroup
	ch := make(chan int, 16)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ch <- fib(15 + i%5)
		}(i)
	}
	go func() { wg.Wait(); close(ch) }()
	sum := 0
	for v := range ch {
		sum += v
	}
	check("goroutines/channels sum="+strconv.Itoa(sum), sum == 20*(610+987+1597+2584+4181))

	// mutex contention
	var mu sync.Mutex
	n := 0
	var wg2 sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			for j := 0; j < 1000; j++ {
				mu.Lock()
				n++
				mu.Unlock()
			}
		}()
	}
	wg2.Wait()
	check("mutex", n == 8000)

	// GC: allocate lots of garbage
	var keep [][]byte
	for i := 0; i < 2000; i++ {
		b := make([]byte, 4096)
		b[0] = byte(i)
		if i%100 == 0 {
			keep = append(keep, b)
		}
	}
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	check(fmt.Sprintf("GC cycles=%d heap=%dKB", ms.NumGC, ms.HeapAlloc/1024), ms.NumGC > 0 && len(keep) == 20)

	// maps and sort
	m := map[string]int{}
	for i := 0; i < 1000; i++ {
		m["k"+strconv.Itoa(i)] = i
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	check("map/sort", len(keys) == 1000 && keys[0] == "k0" && m["k999"] == 999)

	// floating point (soft float)
	f := math.Sqrt(2) * math.Pi / 3
	s := strconv.FormatFloat(f, 'f', 6, 64)
	check("float "+s, s == "1.480961")
	check("float parse", func() bool { v, _ := strconv.ParseFloat("6.02214076e23", 64); return v == 6.02214076e23 }())
	big := uint64(10000000000000000000)
	check("int64 div", int64(-9e18)/7 == -1285714285714285714 && big%13 == 10000000000000000000%13 && big/3 == 3333333333333333333)

	// crypto
	h := sha256.Sum256([]byte("hello wii"))
	check(fmt.Sprintf("sha256 %x", h[:8]), fmt.Sprintf("%x", h) == "9a0bfd1be7a1e062c92a6e7e4a2fa4e46b3cdc8d25de7f30dcf388a33bcd005b")

	// timers
	t0 := time.Now()
	time.Sleep(50 * time.Millisecond)
	el := time.Since(t0)
	check(fmt.Sprintf("sleep %v (%dns, t0=%v)", el.Round(time.Millisecond), el.Nanoseconds(), t0.Format(time.RFC3339Nano)), el >= 50*time.Millisecond && el < 2*time.Second)
	select {
	case <-time.After(20 * time.Millisecond):
		check("time.After", true)
	case <-time.After(5 * time.Second):
		check("time.After", false)
	}

	// stack growth
	check("stack growth", deep(20000) >= 0)

	// panic/recover and errors
	func() {
		defer func() {
			r := recover()
			check(fmt.Sprintf("recover %v", r), r != nil)
		}()
		var a []int
		_ = a[5]
	}()
	e := fmt.Errorf("wrap: %w", os.ErrNotExist)
	check("errors.Is", errors.Is(e, os.ErrNotExist))

	// interfaces/reflection via fmt
	type pt struct{ X, Y int }
	check("fmt struct", fmt.Sprintf("%+v", pt{1, 2}) == "{X:1 Y:2}")
	check("strings", strings.ToUpper(strings.Repeat("ab", 3)) == "ABABAB")

	fmt.Printf("done in %v, NumCPU=%d, goroutines=%d\n", time.Since(start).Round(time.Millisecond), runtime.NumCPU(), runtime.NumGoroutine())
}
