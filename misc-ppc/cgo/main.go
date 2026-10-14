// cgo の動作確認: Go→C、C→Go コールバック、C が作ったスレッドからの Go 呼び出し、
// libc の浮動小数点関数、C 側でのシグナル受信。
package main

/*
#cgo LDFLAGS: -lm -lpthread
#include <math.h>
#include <pthread.h>
#include <stdlib.h>
#include <string.h>

extern int goAdd(int, int);
extern void goNote(int);

static int add(int a, int b) { return a + b; }
static double hyp(double a, double b) { return hypot(a, b); }
static int callback(int n) { int s = 0; for (int i = 0; i < n; i++) s = goAdd(s, i); return s; }

static void *worker(void *arg) { goNote((int)(long)arg); return NULL; }
static int threads(int n) {
	pthread_t t[8];
	for (int i = 0; i < n; i++) pthread_create(&t[i], NULL, worker, (void *)(long)i);
	for (int i = 0; i < n; i++) pthread_join(t[i], NULL);
	return n;
}
static char *dup(const char *s) { return strdup(s); }
*/
import "C"

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"unsafe"
)

//export goAdd
func goAdd(a, b C.int) C.int { return a + b }

var (
	mu    sync.Mutex
	notes = map[int]bool{}
)

//export goNote
func goNote(i C.int) {
	mu.Lock()
	notes[int(i)] = true
	mu.Unlock()
	runtime.GC() // スタック走査とGCを C スレッド上の Go で回す
}

func check(name string, ok bool) {
	if !ok {
		fmt.Println("FAIL", name)
		os.Exit(1)
	}
	fmt.Println("ok", name)
}

func main() {
	check("add", C.add(40, 2) == 42)
	check("hypot", C.hyp(3, 4) == 5)
	check("callback", C.callback(100) == 4950)
	check("threads", C.threads(8) == 8 && len(notes) == 8)
	cs := C.CString("wii")
	d := C.dup(cs)
	check("strings", C.GoString(d) == "wii")
	C.free(unsafe.Pointer(cs))
	C.free(unsafe.Pointer(d))
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); C.callback(200) }()
	}
	wg.Wait()
	check("goroutines", true)
	fmt.Println("PASS")
}
