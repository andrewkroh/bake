package common

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

var log = logrus.WithField("package", "common")

// FindGitProjectRoot searches up from the CWD to find the root of a git project.
// This should return the same value as `git rev-parse --show-toplevel`.
func FindGitProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		_, err := os.Lstat(filepath.Join(dir, ".git"))
		if err != nil {
			if os.IsNotExist(err) {
				up := filepath.Dir(dir)
				if len(up) == len(dir) {
					return "", errors.New("git project root not found")
				}
				dir = up
				continue
			}
			return "", err
		}

		return dir, nil
	}
}

// RunCommand is a wrapper around exec.Command.Output that provides a more
// descriptive error message on failure then the default provided by ExitError.
func RunCommand(cmd *exec.Cmd) ([]byte, error) {
	out, err := cmd.Output()
	if err != nil {
		fullCmd := strings.Join(cmd.Args, " ")
		if exitError, ok := err.(*exec.ExitError); ok {
			return out, errors.Errorf(`command failed (cmd="%v"): %v (stderr=%v)`, fullCmd, exitError, string(exitError.Stderr))
		}
		return out, errors.Errorf(`command failed (cmd="%v"): %v`, fullCmd, err)
	}
	return out, nil
}

// GoFiles returns a list of non-vendor go files.
func GoFiles() ([]string, error) {
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

// GoPackages return a list of non-vendor go packages.
func GoPackages() ([]string, error) {
	out, err := RunCommand(exec.Command("go", "list", "./..."))
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

// Sha256Sum returns a hex encoded sha256 sum of the specified file.
func Sha256Sum(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", errors.Wrap(err, "failed to open file for sha256 sum")
	}
	defer f.Close()

	h := sha256.New()

	if _, err := io.Copy(h, f); err != nil {
		return "", errors.Wrap(err, "failed to calculate sha256 sum")
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
