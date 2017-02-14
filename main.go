package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if pkgs, err := getAllStdPackages(ctx); err == nil {
		allPackage = pkgs
	} else {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	typos, err := FindTypo(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	for _, typo := range typos {
		fmt.Println(strconv.Quote(typo.Text), "at", typo.Pos)
	}
}
