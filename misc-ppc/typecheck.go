// typecheck は GOARCH=ppc 向けに標準パッケージを型検査し、ppc 実装が欠けている所を一覧する(M0 用)。
package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
)

func main() {
	fset := token.NewFileSet()
	imp := importer.ForCompiler(fset, "source", nil)
	conf := types.Config{Importer: imp.(types.ImporterFrom), Sizes: types.SizesFor("gc", build.Default.GOARCH)}
	failed := 0
	for _, path := range os.Args[1:] {
		p, err := build.Import(path, "", 0)
		if err != nil {
			fmt.Printf("NG %s: %v\n", path, err)
			failed++
			continue
		}
		var files []*ast.File
		for _, f := range p.GoFiles {
			af, err := parser.ParseFile(fset, filepath.Join(p.Dir, f), nil, 0)
			if err != nil {
				panic(err)
			}
			files = append(files, af)
		}
		var errs []error
		conf.Error = func(err error) { errs = append(errs, err) }
		conf.Check(path, fset, files, nil)
		if len(errs) > 0 {
			fmt.Printf("NG %s (%d errors)\n", path, len(errs))
			for _, e := range errs[:min(len(errs), 60)] {
				fmt.Printf("   %v\n", e)
			}
			failed++
		} else {
			fmt.Printf("OK %s\n", path)
		}
	}
	os.Exit(failed)
}
