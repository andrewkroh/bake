package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/andrewkroh/bake/golang"
	"gopkg.in/alecthomas/kingpin.v2"
)

var version = "0.0.10"

var (
	app   = kingpin.New("bake", "Beats make tool for building Elastic Beats.")
	debug = app.Flag("debug", "Enable debug logging").Bool()

	gvm             = app.Command("gvm", "Go version management")
	gvmUseGoVersion = gvm.Flag("project-go", "Use project's Go version").Bool()
	gvmPowershell   = gvm.Flag("powershell", "Output powershell commands (windows only)").Bool()
	gvmVersion      = gvm.Arg("version", "golang version").String()

	info          = app.Command("info", "Project info")
	infoGoVersion = info.Flag("go-version", "Print golang version used by project").Bool()
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
	ver, err := getProjectGoVersion()
	if err != nil {
		return err
	}

	fmt.Println(ver)
	return nil
}

func getProjectGoVersion() (string, error) {
	ver, err := parseVersionsAsciidoc()
	if err == nil {
		return ver, nil
	}
	log.Error(err)

	ver, err = parseTravisYml()
	if err == nil {
		return ver, nil
	}
	log.Error(err)

	return "", fmt.Errorf("failed to detect the project's golang version")
}

func parseVersionsAsciidoc() (string, error) {
	file, err := os.Open("libbeat/docs/version.asciidoc")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 2)
		if len(parts) == 2 && parts[0] == ":go-version:" {
			goVersion := strings.TrimSpace(parts[1])
			return goVersion, nil
		}
	}

	return "", fmt.Errorf("go-version not found")
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
