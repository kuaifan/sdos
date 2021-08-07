package main

import (
	"fmt"
	"github.com/kuaifan/sdos/cmd"
	"os"
)

var (
	BuildTime  string
	CommitSha1 string
)

func main() {
	args := os.Args
	if "-v" == args[1] || "--version" == args[1] {
		fmt.Println("BuildTime:\t", BuildTime)
		fmt.Println("CommitSha1:\t", CommitSha1)
		os.Exit(0)
	}
	cmd.Execute()
}
