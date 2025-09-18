package supervisor

import "os/exec"

type Args = []string
type CtrlKind int

const (
	CtrlRegister CtrlKind = iota
	CtrlUnregister
	CtrlShutdown
)

type ExitEvent struct {
	camID  string
	procID int
	status int
	err    error
}

type CtrlEvent struct {
	kind    CtrlKind
	camUUID string
	args    Args
}

type Proc struct {
	procArgs Args
	cmd      *exec.Cmd
}
