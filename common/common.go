package common

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"bytes"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

var log = logrus.WithField("package", "integ")

var ExecutableExtension = ""

func init() {
	if runtime.GOOS == "windows" {
		ExecutableExtension = ".exe"
	}
}

func DownloadFile(url, destinationDir string) (string, error) {
	log.WithField("url", url).Debug("downloading file")

	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "http get failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("download failed with http status %v", resp.StatusCode)
	}

	name := filepath.Join(destinationDir, filepath.Base(url))
	f, err := os.Create(name)
	if err != nil {
		return "", errors.Wrap(err, "failed to create output file")
	}

	numBytes, err := io.Copy(f, resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to write file to disk")
	}
	log.WithFields(logrus.Fields{"file": name, "size_bytes": numBytes}).Debug("download complete")

	return name, nil
}

func WriteTemplateToFile(t *template.Template, data interface{}, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}

	err = t.Execute(f, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %v", err)
	}

	return nil
}

func RunGoCmd(goroot, gopath, dir string, args ...string) (string, error) {
	cmd := exec.Command(filepath.Join(goroot, "bin", "go"), args...)
	cmd.Dir = dir
	cmd.Env = []string{
		fmt.Sprintf("PATH=%v", os.Getenv("PATH")), // cgo builds need a compiler.
		fmt.Sprintf("GOROOT=%v", goroot),
		fmt.Sprintf("GOPATH=%v", gopath),
	}
	var combinedOutput bytes.Buffer
	cmd.Stdout = &combinedOutput
	cmd.Stderr = &combinedOutput

	log.WithFields(logrus.Fields{
		"command": strings.Join(cmd.Args, " "),
		"dir":     cmd.Dir,
		"env":     cmd.Env,
	}).Debug("running command")

	err := cmd.Run()
	output := combinedOutput.String()
	return output, err
}

func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
