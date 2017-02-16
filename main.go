package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/andrewkroh/bake/golang"
	"gopkg.in/alecthomas/kingpin.v2"
)

var version = "0.0.10"

var (
	app   = kingpin.New("bake", "Utility for working with Go projects")
	debug = app.Flag("debug", "Enable debug logging").Bool()

	gvm             = app.Command("gvm", "Go version management")
	gvmUseGoVersion = gvm.Flag("project-go", "Use project's Go version").Bool()
	gvmPowershell   = gvm.Flag("powershell", "Output powershell commands (windows only)").Bool()
	gvmVersion      = gvm.Arg("version", "golang version").String()

	info           = app.Command("info", "Project info")
	infoProjectGo  = info.Flag("project-go", "Print golang version used by project").Bool()
	infoGoFiles    = info.Flag("go-files", "Print Go files sans vendor").Bool()
	infoGoPackages = info.Flag("go-packages", "Print Go packages sans vendor").Bool()

	check    = app.Command("check", "Run checks on the project")
	checkFmt = check.Arg("fmt", "Check that all code is formatted with gofmt").String()
)

var log = logrus.WithField("package", "main")

func main() {
	kingpin.Version(version)
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	logrus.SetLevel(logrus.DebugLevel)
	if *debug {
		logrus.SetOutput(os.Stderr)
	} else {
		logrus.SetOutput(ioutil.Discard)
	}

	var err error
	switch command {
	case gvm.FullCommand():
		err = doGvm()
	case info.FullCommand():
		err = doInfo()
	case check.FullCommand():
		err = doCheck()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func doGvm() error {
	version := *gvmVersion
	if *gvmUseGoVersion == true {
		ver, err := getProjectGoVersion()
		if err != nil {
			return err
		}
		version = ver
	}

	if version == "" {
		return fmt.Errorf("no version specified")
	}
	log.Debugf("Using go version %v", version)

	goroot, err := golang.SetupGolang(version)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		if *gvmPowershell {
			fmt.Printf(`$env:GOROOT = "%v"`+"\n", goroot)
			fmt.Printf(`$env:PATH = "$env:GOROOT\bin;$env:PATH"` + "\n")
		} else {
			fmt.Printf("set GOROOT=%v\n", goroot)
			fmt.Printf("set PATH=%s\bin;%s\n", goroot, os.Getenv("PATH"))
		}
	} else {
		fmt.Printf(`export GOROOT="%v"`+"\n", goroot)
		fmt.Println(`export PATH="$GOROOT/bin:$PATH"`)
	}

	if strings.HasPrefix(version, "1.5") {
		if runtime.GOOS == "windows" {
			fmt.Println("set GO15VENDOREXPERIMENT=1")
		} else {
			fmt.Println("export GO15VENDOREXPERIMENT=1")
		}
	}

	return nil
}

func doInfo() error {
	switch {
	default:
		return fmt.Errorf("no info flag specified")
	case *infoProjectGo:
		ver, err := getProjectGoVersion()
		if err != nil {
			return err
		}
		fmt.Println(ver)
	case *infoGoFiles:
		files, err := goFiles()
		if err != nil {
			return err
		}
		for _, f := range files {
			fmt.Println(f)
		}
	case *infoGoPackages:
		packages, err := goPackages()
		if err != nil {
			return err
		}
		for _, f := range packages {
			fmt.Println(f)
		}
	}
	return nil
}

func doCheck() error {
	switch {
	default:
		return fmt.Errorf("no check argument specified")
	case *checkFmt != "":
		files, err := checkFormatting()
		if err != nil {
			return err
		}

		if len(files) > 0 {
			for _, f := range files {
				fmt.Println(f)
			}
			return fmt.Errorf("these files need to be formatted with gofmt")
		}
	}
	return nil
}

func checkFormatting() ([]string, error) {
	files, err := goFiles()
	if err != nil {
		return nil, err
	}

	args := make([]string, 0, len(files)+2)
	args = append(args, "-s", "-l")
	args = append(args, files...)

	out, err := exec.Command("gofmt", args...).Output()
	if err != nil {
		return nil, err
	}

	return strings.Fields(string(out)), nil
}

func getProjectGoVersion() (string, error) {
	ver, err := parseTravisYml()
	if err != nil {
		return "", fmt.Errorf("failed to detect the project's golang version: %v", err)
	}

	return ver, nil
}

func parseTravisYml() (string, error) {
	filename := ".travis.yml"
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	var re = regexp.MustCompile(`(?mi)^go:\r?\n\s*-\s+(\S+)\s*$`)
	matches := re.FindAllStringSubmatch(string(file), 1)
	if len(matches) == 0 {
		return "", fmt.Errorf("go not found in %v", filename)
	}

	goVersion := matches[0][1]
	return goVersion, nil
}

func goFiles() ([]string, error) {
	var files []string
	callback := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		// Select .go files only.
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Filter vendor
		for _, dir := range strings.Split(path, string(filepath.Separator)) {
			if dir == "vendor" {
				return nil
			}
		}

		files = append(files, path)
		return nil
	}

	return files, filepath.Walk(".", callback)
}

func goPackages() ([]string, error) {
	out, err := exec.Command("go", "list", "./...").Output()
	if err != nil {
		return nil, err
	}

	packages := strings.Split(string(out), "\n")
	filtered := packages[:0]

	// Filter vendor
outer:
	for _, p := range packages {
		for _, dir := range strings.Split(p, string(filepath.Separator)) {
			if dir == "vendor" {
				continue outer
			}
		}

		if strings.TrimSpace(p) != "" {
			filtered = append(filtered, p)
		}
	}

	return filtered, nil
}
