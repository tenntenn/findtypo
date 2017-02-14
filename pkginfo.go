package main

import (
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
)

type PkgInfo struct {
	Pkg  *build.Package
	Info *types.Info
}

func (p *PkgInfo) HasIdent(s string) bool {
	for _, o := range p.Info.Defs {
		if o == nil {
			continue
		}

		if o.Exported() && o.Name() == s {
			return true
		}
	}
	return false
}

// すべての標準パッケージを取得する
func getAllStdPackages(ctx *build.Context) (map[string][]*PkgInfo, error) {
	allpkgs := map[string][]*PkgInfo{}

	config := &types.Config{
		Importer: importer.Default(),
	}

	srcDir := filepath.Join(ctx.GOROOT, "src")
	err := filepath.Walk(srcDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ディレクトリ意外は無視
		if !fi.IsDir() {
			return nil
		}

		// internal, testdata, vendorディレクトリは無視する
		if n := fi.Name(); n == "internal" || n == "testdata" || n == "vendor" {
			return filepath.SkipDir
		}

		// ディレクトリ単位でインポートする
		pkg, err := ctx.ImportDir(path, 0)
		if err != nil {
			return nil
		}

		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, pkg.Dir, makeFilter(pkg), 0)
		if err != nil {
			return nil
		}

		for pn, p := range pkgs {
			// 必要なのはそのパッケージで定義された型のみ
			info := &types.Info{
				Defs: map[*ast.Ident]types.Object{},
			}

			// マップをスライスに変換する
			files := make([]*ast.File, 0, len(p.Files))
			for _, f := range p.Files {
				files = append(files, f)
			}

			// 型チェックを行う
			_, err := config.Check(pn, fset, files, info)
			if err != nil {
				return nil
			}

			allpkgs[pn] = append(allpkgs[pn], &PkgInfo{
				Pkg:  pkg,
				Info: info,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return allpkgs, nil
}
