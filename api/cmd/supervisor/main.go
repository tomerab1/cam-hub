package supervisor

import (
	visor "tomerab.com/cam-hub/internal/supervisor"
)

func main() {
	supervisor := visor.NewSupervisor(10)
	supervisor.Run()
}
