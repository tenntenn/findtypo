package main

import (
	"go/build"
	"go/types"
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
	return allpkgs, nil
}
