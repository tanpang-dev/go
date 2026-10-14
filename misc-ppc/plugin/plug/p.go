package main

import (
	"fmt"
	"strings"
	"sync"
)

var Counter int

func Add(a, b int) int { return a + b }
func Inc() int        { Counter++; return Counter }

type greeter struct{ prefix string }

func (g greeter) Greet(s string) string { return g.prefix + strings.ToUpper(s) }
func (g greeter) String() string        { return "greeter" }

func NewGreeter() any { return greeter{"hello, "} }

func Work(n int) int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	total := 0
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m := map[int]string{}
			for j := 0; j < 1000; j++ {
				m[j] = fmt.Sprint(i + j)
			}
			mu.Lock()
			total += len(m)
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	return total
}
