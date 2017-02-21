package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/andrewkroh/bake/common"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

const (
	dockerComposeCmd = "docker-compose"
)

var dockerLog = logrus.WithField("package", "main").WithField("cmd", "docker")

func registerDockerCommand(app *kingpin.Application) {
	cmd := &DockerCommand{}
	docker := app.Command("docker", "Start test services powered by Docker "+
		"and open a shell on the host where environment variables point to services.").Action(cmd.Run)
	docker.Flag("project", "Specify an alternate project name (default: directory name)").Short('p').StringVar(&cmd.Project)
	docker.Flag("file", "Specify an alternate compose file (default: docker-compose.yml)").Short('f').Default("docker-compose.yml").StringsVar(&cmd.Files)
	docker.Flag("log", "Specify log output file").Short('o').StringVar(&cmd.Log)
	docker.Arg("script", "Script to run").StringVar(&cmd.Script)
}

type DockerCommand struct {
	Project string
	Files   []string
	Log     string
	Script  string
}

func (c *DockerCommand) Run(ctx *kingpin.ParseContext) error {
	var args []string
	if c.Project != "" {
		args = append(args, []string{"-p", c.Project}...)
	}
	for _, f := range c.Files {
		args = append(args, []string{"-f", f}...)
	}

	configYAML, err := common.RunCommand(exec.Command(dockerComposeCmd, append(args, "config")...))
	if err != nil {
		return errors.Wrap(err, "failed to get docker-compose config")
	}

	cmd, err := c.dockerComposeUp(args)
	if err != nil {
		return err
	}
	defer cmd.SendCtrlCSignal()

	env, err := getServicePorts(args, configYAML)
	if err != nil {
		return err
	}

	return shell(env)
}

func (c *DockerCommand) dockerComposeUp(args []string) (*common.Cmd, error) {
	cmd := common.Command(dockerComposeCmd, append(args, "up")...)
	cmd.Stdout = ioutil.Discard

	var logFile *os.File
	if c.Log != "" {
		var err error
		logFile, err = os.Create(c.Log)
		if err != nil {
			return nil, err
		}
		cmd.Stdout = logFile
	}

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			logFile.Close()
		}
		return nil, err
	}

	go func() {
		cmd.Wait()
		if logFile != nil {
			logFile.Close()
		}
	}()

	return cmd, nil
}

func shell(env map[string]string) error {
	envVars := os.Environ()
	for k, v := range env {
		envVars = append(envVars, fmt.Sprintf("%v=%v", k, v))
	}

	cmd := &exec.Cmd{
		Path:   "/bin/bash",
		Env:    envVars,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	return cmd.Run()
}

type Config struct {
	Services map[string]Service
}

type Service struct {
	Ports []string
}

func getServicePorts(fileArgs []string, config []byte) (map[string]string, error) {
	c := Config{}
	err := yaml.Unmarshal(config, &c)
	if err != nil {
		return nil, err
	}

	env := map[string]string{}
	for name, service := range c.Services {
		for _, port := range service.Ports {
			upperServiceName := strings.ToUpper(name)
			hostKey := fmt.Sprintf("%s_HOST", upperServiceName)
			portKey := fmt.Sprintf("%s_PORT_%s_TCP_PORT", upperServiceName, port)

			host, mappedPort, err := getPortMapping(fileArgs, name, port)
			if err != nil {
				dockerLog.WithError(err).Error("service will be unavailable")
			}

			env[hostKey] = host
			env[portKey] = mappedPort
		}
	}
	return env, nil
}

func getPortMapping(fileArgs []string, service, port string) (string, string, error) {
	mapping, err := exec.Command(dockerComposeCmd, append(fileArgs, "port", service, port)...).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", "", errors.Wrapf(err, "failed to get docker-compose port mapping for %s:%s (%v)", service, port, string(bytes.TrimSpace(exitErr.Stderr)))
		}
		return "", "", errors.Wrapf(err, "failed to get docker-compose port mapping for %s:%s", service, port)
	}

	mappedHostPort := string(bytes.TrimSpace(mapping))

	dockerLog.Infof("service %v %v->%v", service, mappedHostPort, port)
	host, port, err := net.SplitHostPort(mappedHostPort)
	if err != nil {
		return "", "", err
	}

	if ip := net.ParseIP(host); ip != nil {
		switch {
		case ip.Equal(net.IPv4zero):
			host = "127.0.0.1"
		case ip.Equal(net.IPv6zero):
			host = net.IPv6loopback.String()
		}
	}

	return host, port, nil
}
