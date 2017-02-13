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
		if err := doGvm(); err != nil {
			fmt.Fprintf(os.Stderr, "Errof: %v\n", err)
			os.Exit(1)
		}
	}
}

func doGvm() error {
	goroot, err := golang.SetupGolang(*gvmVersion)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		fmt.Printf(`set GOROOT="%v"`+"\n", goroot)
		fmt.Println(`set PATH="%GOROOT%\bin:%PATH%"`)
	} else {
		fmt.Printf(`export GOROOT="%v"`+"\n", goroot)
		fmt.Println(`export PATH="$GOROOT/bin:$PATH"`)
	}

	return nil
}
