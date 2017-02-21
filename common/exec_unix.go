// +build !windows

package common

import (
	"context"
	"os"
	"os/exec"
)

func Command(name string, arg ...string) *Cmd {
	cmd := exec.Command(name, arg...)
	return &Cmd{*cmd}
}

func CommandContext(ctx context.Context, name string, arg ...string) *Cmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	return &Cmd{*cmd}
}

func (c *Cmd) SendCtrlCSignal() error {
	log.WithField("pid", c.Process.Pid).Debug("sending ctrl+C to process")
	return c.Process.Signal(os.Interrupt)
}
