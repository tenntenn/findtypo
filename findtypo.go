package main

import (
	"go/token"
	"os"
)

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
	return nil, nil
}
