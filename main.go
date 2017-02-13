package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/andrewkroh/bake/golang"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("bake", "Beats make tool for building Elastic Beats.")

	gvm        = app.Command("gvm", "Go version management")
	gvmVersion = gvm.Arg("version", "golang version").Required().String()
)

func main() {
	kingpin.Version("0.0.1")
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// Post message
	case gvm.FullCommand():
		if err := gvm(); err != nil {
			fmt.Fprintf(os.Stderr, "Errof: %v\n", err)
			os.Exit(1)
		}
	}
}

func gvm() error {
	goroot, err := golang.SetupGolang(*gvmVersion)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		fmt.Printf(`GOROOT="%v"\n`, goroot)
		fmt.Printf(`PATH="%GOROOT%\bin:$PATH"\n`)
	} else {
		fmt.Printf(`GOROOT="%v"\n`, goroot)
		fmt.Printf(`PATH="$GOROOT/bin:$PATH"\n`)
	}
}
