package common

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procGenerateConsoleCtrlEvent = modkernel32.NewProc("GenerateConsoleCtrlEvent")
)

func Command(name string, arg ...string) *Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	return &Cmd{*cmd}
}

func CommandContext(ctx context.Context, name string, arg ...string) *Cmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	return &Cmd{*cmd}
}

func (c *Cmd) SendCtrlCSignal() error {
	if c.SysProcAttr == nil || c.SysProcAttr.CreationFlags&syscall.CREATE_NEW_PROCESS_GROUP == 0 {
		// If the CREATE_NEW_PROCESS_GROUP wasn't set then sending the signal
		// will end up kill this process too since they would all be in the
		// same process group.
		return fmt.Errorf("usage error on Windows: Cmd.SysProcAttr.CreationFlags must have CREATE_NEW_PROCESS_GROUP set")
	}
	log.WithField("pid", c.Process.Pid).Debug("sending ctrl+C to process")
	return sendCtrlBreak(c.Process.Pid)
}

func sendCtrlBreak(pid int) error {
	err := GenerateConsoleCtrlEvent(syscall.CTRL_BREAK_EVENT, uint32(pid))
	return fmt.Errorf("failed to GenerateConsoleCtrlEvent: %v", err)
}

func GenerateConsoleCtrlEvent(event uint32, processgroupid uint32) (err error) {
	r1, _, e1 := syscall.Syscall(procGenerateConsoleCtrlEvent.Addr(), 2, uintptr(event), uintptr(processgroupid), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
