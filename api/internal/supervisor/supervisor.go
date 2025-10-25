package supervisor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type Supervisor struct {
	mtx    sync.Mutex
	procs  map[string]*Proc
	exitCh chan ExitEvent
	ctrlCh chan CtrlEvent
	logger *slog.Logger
}

func NewSupervisor(maxProcs int, logger *slog.Logger) *Supervisor {
	return &Supervisor{
		mtx:    sync.Mutex{},
		procs:  make(map[string]*Proc),
		exitCh: make(chan ExitEvent, maxProcs),
		ctrlCh: make(chan CtrlEvent, maxProcs),
		logger: logger,
	}
}

func (visor *Supervisor) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case exit := <-visor.exitCh:
			visor.logger.Info("Process exit",
				"cam", exit.camID,
				"pid", exit.procID,
				"status", exit.status,
				"err", exit.err,
			)
		case ev := <-visor.ctrlCh:
			switch ev.Kind {
			case CtrlRegister:
				visor.Register(ev.CamUUID, ev.Args)
			case CtrlUnregister:
				visor.Unregister(ev.CamUUID)
			case CtrlShutdown:
				return
			}
		}
	}
}

// Returns the revison of the camera.
func (visor *Supervisor) GetCameraRevision(camUUID string) (int, error) {
	p := visor.findProc(camUUID)
	if p == nil {
		return -1, fmt.Errorf("camera (%s) is not registered", camUUID)
	}

	return p.Version, nil
}

func (visor *Supervisor) Shutdown() {
	visor.logger.Info("Shutting down")
	for uuid := range visor.procs {
		visor.Unregister(uuid)
	}
}

func (visor *Supervisor) NotifyCtrl(ev CtrlEvent) {
	visor.ctrlCh <- ev
}

func (visor *Supervisor) Register(camUUID string, args Args) {
	if visor.findProc(camUUID) != nil {
		visor.logger.Info("register: process for camera is already running", "uuid", camUUID)
		return
	}

	cmd := exec.Command("go", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := cmd.Start()
	if err != nil {
		visor.exitCh <- ExitEvent{
			camID:  camUUID,
			procID: -1,
			status: -1,
			err:    err,
		}
		return
	}
	visor.logger.Info("started new process", "pid", cmd.Process.Pid, "args", args)

	visor.mtx.Lock()
	defer visor.mtx.Unlock()
	visor.procs[camUUID] = &Proc{
		procArgs: args,
		cmd:      cmd,
		Version:  1,
	}

	go func() {
		err := cmd.Wait()
		visor.exitCh <- ExitEvent{
			camID:  camUUID,
			procID: cmd.Process.Pid,
			status: cmd.ProcessState.ExitCode(),
			err:    err,
		}

		if visor.findProc(camUUID) == nil {
			return
		}

		ok := visor.deleteProc(camUUID)
		if !ok {
			visor.logger.Error("register: wait: failed to delete the camera", "uuid", camUUID)
			return
		}
	}()
}

func (visor *Supervisor) Unregister(camUUID string) {
	proc := visor.findProc(camUUID)
	if proc == nil {
		visor.logger.Error("unregister: failed to find process", "uuid", camUUID)
		return
	}
	visor.logger.Debug("unregister", "camUUID", camUUID)

	p := proc.cmd.Process
	if err := syscall.Kill(-p.Pid, syscall.SIGTERM); err != nil {
		if err := syscall.Kill(-p.Pid, syscall.SIGKILL); err != nil {
			visor.logger.Error("unregister: kill failed", "uuid", camUUID, "err", err)
		}
	} else {
		go func(p *os.Process) {
			time.Sleep(5 * time.Second)
			// Check if the process is still alive, if it is send sigkill
			if err := syscall.Kill(-p.Pid, 0); err == nil {
				if err := syscall.Kill(-p.Pid, syscall.SIGKILL); err != nil {
					visor.logger.Error("unregister: kill failed", "uuid", camUUID, "err", err)
				}
			}
		}(p)
	}
}

func (visor *Supervisor) deleteProc(camUUID string) bool {
	visor.mtx.Lock()
	defer visor.mtx.Unlock()

	_, ok := visor.procs[camUUID]
	if !ok {
		return false
	}

	delete(visor.procs, camUUID)
	return true
}

func (visor *Supervisor) findProc(camUUID string) *Proc {
	visor.mtx.Lock()
	defer visor.mtx.Unlock()
	proc, ok := visor.procs[camUUID]
	if !ok {
		return nil
	}

	return proc
}
