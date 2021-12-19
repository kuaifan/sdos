package main

import (
	"github.com/kuaifan/sdos/cmd"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()
}
