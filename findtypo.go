package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

var (
	ctx *build.Context
)

func init() {
	c := build.Default // copy
	ctx = &c
	ctx.CgoEnabled = false
	ctx.GOPATH = ""
}

type Typo struct {
	Pos  token.Position
	Text string
}

func FindTypo(paths []string) ([]*Typo, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var typos []*Typo
	for _, path := range paths {
		t, err := findTypo(dir, path)
		if err != nil {
			return nil, err
		}
		typos = append(typos, t...)
	}

	return typos, nil
}

func findTypo(dir string, path string) ([]*Typo, error) {
	pkg, err := ctx.Import(path, dir, build.IgnoreVendor)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkg.Dir, makeFilter(pkg), parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var typos []*Typo
	for _, p := range pkgs {
		for _, f := range p.Files {
			t, err := findTypoByFile(fset, f)
			if err != nil {
				return nil, err
			}
			typos = append(typos, t...)
		}
	}

	return typos, nil
}

// 不要なファイルを省く
func makeFilter(pkg *build.Package) func(fi os.FileInfo) bool {
	return func(fi os.FileInfo) bool {
		// テストファイルを省く
		if strings.HasSuffix(fi.Name(), "_test.go") {
			return false
		}

		// 無視されているファイルを省く
		for _, ignored := range pkg.IgnoredGoFiles {
			if ignored == fi.Name() {
				return false
			}
		}

		// CGO関係のファイルを省く
		for _, cgofile := range pkg.CgoFiles {
			if cgofile == fi.Name() {
				return false
			}
		}

		return true
	}
}

// ファイルごとにタイポを見つける
func findTypoByFile(fset *token.FileSet, f *ast.File) ([]*Typo, error) {
	var typos []*Typo

	for _, cg := range f.Comments {
		s := bufio.NewScanner(strings.NewReader(cg.Text()))
		s.Split(bufio.ScanWords)
		for s.Scan() {
			str := strings.TrimRight(s.Text(), ".")
			if typo, isTarget := hasTypo(str); typo && isTarget {

				// 末尾のsを取り除いてもう一度やってみる
				if strings.HasSuffix(str, "s") {
					str = strings.TrimRight(str, "s")
					if typo, isTarget := hasTypo(str); !typo && isTarget {
						// sを取り除いたら大丈夫だった
						continue
					}
				}

				typos = append(typos, &Typo{
					Text: s.Text(),
					Pos:  fset.Position(cg.Pos()),
				})
			}

		}

		if err := s.Err(); err != nil {
			return nil, err
		}
	}

	return typos, nil
}

func hasTypo(s string) (typo, target bool) {
	expr, err := parser.ParseExpr(s)
	if err != nil {
		return false, false
	}

	pkg, ident, ok := getPkgIdent(expr)
	if !ok {
		return false, false
	}

	// パッケージが存在するか
	if !isExistPkg(pkg) {
		return false, false
	}

	// 識別子が存在するか
	typo = !isExitIdent(pkg, ident)

	return typo, true
}

// パッケージ名と識別子を取得する
func getPkgIdent(expr ast.Expr) (pkg, ident string, ok bool) {
	selExpr, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}

	if !selExpr.Sel.IsExported() {
		return "", "", false
	}

	pkgIdent, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return "", "", false
	}

	pkg = pkgIdent.Name
	ident = selExpr.Sel.Name

	return pkg, ident, true
}

func isExistPkg(pkg string) bool {
	fmt.Println("pkg:", pkg)
	return true
}

func isExitIdent(pkg, ident string) bool {
	fmt.Println("ident:", ident)
	return true
}
