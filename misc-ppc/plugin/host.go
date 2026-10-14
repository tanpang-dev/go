// plugin の動作確認: 読み込み、関数・変数の参照、型の共有、goroutine と GC。
package main

import (
	"fmt"
	"os"
	"plugin"
	"runtime"
	"strings"
)

type Greeter interface{ Greet(string) string }

func main() {
	p, err := plugin.Open(os.Args[1])
	if err != nil {
		fmt.Println("FAIL open:", err)
		os.Exit(1)
	}
	ok := true
	f, _ := p.Lookup("Add")
	ok = ok && f.(func(int, int) int)(40, 2) == 42
	v, _ := p.Lookup("Counter")
	*v.(*int) = 7
	inc, _ := p.Lookup("Inc")
	ok = ok && inc.(func() int)() == 8
	g, _ := p.Lookup("NewGreeter")
	gr := g.(func() any)().(Greeter)
	ok = ok && gr.Greet("wii") == "hello, WII"
	w, _ := p.Lookup("Work")
	ok = ok && w.(func(int) int)(50) == 50*1000
	runtime.GC()
	ok = ok && strings.Contains(fmt.Sprint(gr), "greeter")
	if ok {
		fmt.Println("PASS")
	} else {
		fmt.Println("FAIL")
		os.Exit(1)
	}
}
