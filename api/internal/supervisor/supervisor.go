package supervisor

import (
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

func NewSupervisor(maxProcs int) *Supervisor {
	return &Supervisor{
		mtx:    sync.Mutex{},
		procs:  make(map[string]*Proc),
		exitCh: make(chan ExitEvent, maxProcs),
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (visor *Supervisor) Run() {
	for {
		select {
		case exit := <-visor.exitCh:
			visor.logger.Info("Process exit",
				"cam", exit.camID,
				"pid", exit.procID,
				"status", exit.status,
				"err", exit.err,
			)
		case ev := <-visor.ctrlCh:
			switch ev.kind {
			case CtrlRegister:
				visor.Register(ev.camUUID, ev.args)
			case CtrlUnregister:
				visor.Unregister(ev.camUUID)
			case CtrlShutdown:
				visor.Shutdown()
			}
		}
	}
}

func (visor *Supervisor) Shutdown() {
	visor.logger.Info("Shutting down")
}

func (visor *Supervisor) Register(camUUID string, args Args) {
	if visor.findProc(camUUID) != nil {
		visor.logger.Info("register: process for camera is already running", "uuid", camUUID)
		return
	}

	cmd := exec.Command("go", args...)

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

	visor.mtx.Lock()
	defer visor.mtx.Unlock()
	visor.procs[camUUID] = &Proc{
		procArgs: args,
		cmd:      cmd,
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
	p := proc.cmd.Process

	err := p.Signal(syscall.SIGTERM)
	if err != nil {
		visor.logger.Error("unregister: SIGTERM failed", "err", err)
		if err := p.Kill(); err != nil {
			visor.logger.Error("unregister: kill failed", "uuid", camUUID, "err", err)
		}
	} else {
		go func(p *os.Process) {
			time.Sleep(5 * time.Second)
			// Check if the process is still alive, if it is send sigkill
			if p.Signal(syscall.Signal(0)) == nil {
				if err := p.Kill(); err != nil {
					visor.logger.Error("unregister: kill failed", "uuid", camUUID, "err", err)
				}
			}
		}(p)
	}

	visor.deleteProc(camUUID)
	visor.exitCh <- ExitEvent{
		camID:  camUUID,
		procID: p.Pid,
		status: -1,
		err:    err,
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
