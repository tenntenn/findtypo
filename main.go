package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/types"
	"os"
	"strings"

	"github.com/gertd/go-pluralize"
	"golang.org/x/tools/go/packages"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

var stdpkgs []*packages.Package

func run() error {
	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes,
	}

	pkgs, err := packages.Load(cfg, "std")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}

	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	stdpkgs = pkgs
	for _, pkg := range pkgs {
		if err := check(pkg); err != nil {
			return err
		}
	}

	return err
}

func check(pkg *packages.Package) error {
	for _, f := range pkg.Syntax {
		fname := pkg.Fset.File(f.Pos()).Name()
		if strings.HasSuffix(fname, "_test.go") {
			continue
		}
		for _, decl := range f.Decls {
			var cg *ast.CommentGroup
			switch decl := decl.(type) {
			case *ast.GenDecl:
				cg = decl.Doc
			case *ast.FuncDecl:
				cg = decl.Doc
			}
			if cg == nil {
				continue
			}

			s := bufio.NewScanner(strings.NewReader(cg.Text()))
			s.Split(bufio.ScanWords)
			for s.Scan() {
				str := strings.TrimRight(s.Text(), ".")
				expr, err := parser.ParseExpr(str)
				if err != nil {
					continue
				}

				sel, _ := expr.(*ast.SelectorExpr)
				if sel == nil || !sel.Sel.IsExported() {
					continue
				}

				var buf bytes.Buffer
				types.WriteExpr(&buf, sel.X)
				if isTypo(buf.String(), sel.Sel.Name) {
					pos := pkg.Fset.Position(cg.Pos())
					fmt.Println(pos, str)
				}
			}
		}
	}
	return nil
}

func isTypo(pkgname string, sel string) bool {
	var pkg *packages.Package
	for _, p := range stdpkgs {
		if p.ID == pkgname {
			pkg = p
			break
		}
	}

	if pkg == nil {
		return false
	}

	if obj := pkg.Types.Scope().Lookup(sel); obj != nil {
		return false
	}

	pluralize := pluralize.NewClient()
	rmS := pluralize.Singular(sel)
	if obj := pkg.Types.Scope().Lookup(rmS); obj != nil {
		return false
	}

	return true
}
