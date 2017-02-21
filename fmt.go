package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/andrewkroh/bake/common"
	"gopkg.in/alecthomas/kingpin.v2"
)

func registerFmtCommand(app *kingpin.Application) {
	cmd := &FmtCommand{}
	app.Command("fmt", "Run gofmt -s on non-vendor Go files").Action(cmd.Run)
}

type FmtCommand struct{}

func (c *FmtCommand) Run(ctx *kingpin.ParseContext) error {
	files, err := goFmtSimplify()
	if err != nil {
		return err
	}

	for _, f := range files {
		fmt.Println(f)
	}

	return nil
}

// goFmtSimplify run "gofmt -s" on all non-vendor go files in and below the
// current directory.
func goFmtSimplify() ([]string, error) {
	files, err := common.GoFiles()
	if err != nil {
		return nil, err
	}

	args := make([]string, 0, len(files)+3)
	args = append(args, "-s", "-w", "-l")
	args = append(args, files...)

	out, err := common.RunCommand(exec.Command("gofmt", args...))
	if err != nil {
		return nil, err
	}

	return strings.Fields(string(out)), nil
}
