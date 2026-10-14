// c-archive / c-shared の動作確認用ライブラリ。
package main

import "C"

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

//export GoAdd
func GoAdd(a, b C.int) C.int { return a + b }

//export GoUpper
func GoUpper(s *C.char) *C.char { return C.CString(strings.ToUpper(C.GoString(s))) }

//export GoWork
func GoWork(n C.int) C.long {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var sum int64
	for i := 0; i < int(n); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m := map[int]string{}
			for j := 0; j < 1000; j++ {
				m[j] = fmt.Sprint(i * j)
			}
			mu.Lock()
			sum += int64(len(m))
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	runtime.GC()
	return C.long(sum)
}

func main() {}
