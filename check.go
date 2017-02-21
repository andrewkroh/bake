package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/andrewkroh/bake/common"
	"github.com/joeshaw/multierror"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	format = "fmt"
	vet    = "vet"
	notice = "notice"
)

var allChecks = []string{format, vet, notice}

var checkLog = logrus.WithField("package", "main").WithField("cmd", "check")

func registerCheckCommand(app *kingpin.Application) {
	cmd := &CheckCommand{}
	check := app.Command("check", "Run checks on the project. The options are fmt, vet, and notice. By default all checks are run.").Action(cmd.Run)
	check.Arg("checks", "checks to run").EnumsVar(&cmd.Checks, allChecks...)
}

type CheckCommand struct {
	Checks []string
}

func (c *CheckCommand) Run(ctx *kingpin.ParseContext) error {
	checkLog.WithField("cmd", c).Debug("Running checks")

	checks := c.Checks
	if len(c.Checks) == 0 {
		checks = allChecks
	}

	var errs multierror.Errors
	for _, c := range checks {
		checkLog.Debugf("Running %v check", c)

		switch c {
		case format:
			files, err := checkFormatting()
			if err != nil {
				errs = append(errs, err)
				continue
			}
			for _, f := range files {
				fmt.Println(f)
			}
			if len(files) > 0 {
				errs = append(errs, errors.New("some files need to be formatted with gofmt -s"))
			}
		case vet:
			problems, err := checkVet()
			if err != nil {
				errs = append(errs, err)
				continue
			}
			for _, f := range problems {
				fmt.Println(f)
			}
			if len(problems) > 0 {
				errs = append(errs, errors.New("some files have go vet errors"))
			}
		case notice:
			if err := checkNotice(); err != nil {
				errs = append(errs, err)
				continue
			}
		}
	}

	if len(errs) == 1 {
		return errs[0]
	}

	return errs.Err()
}

func checkFormatting() ([]string, error) {
	files, err := common.GoFiles()
	if err != nil {
		return nil, err
	}

	args := make([]string, 0, len(files)+2)
	args = append(args, "-s", "-l")
	args = append(args, files...)

	out, err := common.RunCommand(exec.Command("gofmt", args...))
	if err != nil {
		return nil, err
	}

	return strings.Fields(string(out)), nil
}

func checkVet() ([]string, error) {
	packages, err := common.GoPackages()
	if err != nil {
		return nil, err
	}

	args := make([]string, 0, len(packages)+1)
	args = append(args, "vet")
	args = append(args, packages...)

	out, err := common.RunCommand(exec.Command("go", args...))
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, nil
	}

	return strings.Split(string(out), "\n"), nil
}

func checkNotice() error {
	file := filepath.Join(os.TempDir(), "NOTICE-"+strconv.Itoa(rand.Int()))
	defer os.Remove(file)

	cmd := getNoticeCommandDefaults()
	defaultOutput := cmd.Output
	cmd.Output = file
	if err := generateNotice(cmd); err != nil {
		return err
	}

	existingSum, err := common.Sha256Sum(defaultOutput)
	if err != nil {
		return errors.Wrap(err, "failed reading existing NOTICE file")
	}

	newSum, err := common.Sha256Sum(file)
	if err != nil {
		return errors.Wrap(err, "failed reading new NOTICE file")
	}
	checkLog.WithFields(logrus.Fields{
		"notice_sha256": existingSum,
		"new_sha256":    newSum,
	}).Info("calculated sha256 file sums")

	if existingSum != newSum {
		return errors.Errorf("NOTICE file needs to be updated")
	}
	return nil
}
