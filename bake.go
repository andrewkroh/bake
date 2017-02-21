package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/andrewkroh/bake/common"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app   = kingpin.New("bake", "Utility for working with Beats projects")
	debug = app.Flag("debug", "Enable debug logging").Short('d').Bool()

	// TODO: The following commands are not implemented.

	test         = app.Command("test", "Run tests.").Alias("unit").Alias("integ").Alias("system")
	testCover    = test.Flag("cover", "Generate code coverage output and HTML report").Bool()
	testRace     = test.Flag("race", "Enable race detector while testing").Bool()
	testJUnit    = test.Flag("junit", "Generate JUnit XML report summarizing test results").Bool()
	testTypes    = test.Flag("tests", "Test types to execute. Options are unit (default), benchmark, integ, and system.").Default("unit").Enums("unit", "integ", "system", "benchmark")
	testPackages = test.Arg("packages", "Packages to build. Defaults to building.").Default(".").String()

	crosscompile = app.Command("crosscompile", "Cross-compile the beat without CGO")

	docs = app.Command("docs", "Build the Elastic asciidoc book for the Beat")

	ci = app.Command("ci", "Run all checks and tests.")
)

var (
	// CWD is the current working directory when the process was started.
	CWD string

	// ProjectRootAbs is the absolute path to the root of the project based on
	// the location of the .git directory.
	ProjectRootAbs string

	// ProjectRootRel is the relative path to the root of the project based on
	// the location of the .git directory.
	ProjectRootRel string
)

func init() {
	var err error
	CWD, err = os.Getwd()
	if err != nil {
		app.Fatalf("%v: Failed to determine your current working directory.", err)
	}

	ProjectRootAbs, err = common.FindGitProjectRoot()
	if err != nil {
		app.Fatalf("%v: Are you running bake from within a Git clone?", err)
	}

	ProjectRootRel, err = filepath.Rel(CWD, ProjectRootAbs)
	if err != nil {
		app.Fatalf("%v: Failed to determine the relative path to the Git project root.", err)
	}
}

func main() {
	registerCheckCommand(app)
	registerFmtCommand(app)
	registerNoticeCommand(app)
	registerDockerCommand(app)

	app.HelpFlag.Short('h')
	app.DefaultEnvars()
	app.UsageTemplate(kingpin.SeparateOptionalFlagsUsageTemplate)
	app.PreAction(func(ctx *kingpin.ParseContext) error {
		logrus.SetLevel(logrus.DebugLevel)
		if *debug {
			logrus.SetOutput(os.Stderr)
		} else {
			logrus.SetOutput(ioutil.Discard)
		}
		return nil
	})

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		app.Errorf("%v\n", err)
		os.Exit(1)
	}
}
