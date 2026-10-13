package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	fmt.Printf("hello from Go %s on %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if len(os.Args) > 1 {
		fmt.Println("args:", os.Args[1:])
	}
}
