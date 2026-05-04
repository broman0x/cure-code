package agent

import (
	"os"
	"os/exec"
	"sync"
)

type BackgroundProcess struct {
	ID      int
	Command string
	Cmd     *exec.Cmd
}

// [EN] ProcessManager handles the lifecycle of background processes started by the agent.
// [ID] ProcessManager mengelola siklus hidup proses latar belakang yang dimulai oleh agen.
type ProcessManager struct {
	mu        sync.RWMutex
	processes map[int]*BackgroundProcess
	nextID    int
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[int]*BackgroundProcess),
		nextID:    1,
	}
}

// [EN] Add registers a new background process and returns its tracking ID.
// [ID] Add mendaftarkan proses latar belakang baru dan mengembalikan ID pelacakannya.
func (pm *ProcessManager) Add(command string, cmd *exec.Cmd) int {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	id := pm.nextID
	pm.processes[id] = &BackgroundProcess{
		ID:      id,
		Command: command,
		Cmd:     cmd,
	}
	pm.nextID++
	return id
}

func (pm *ProcessManager) List() []*BackgroundProcess {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	res := make([]*BackgroundProcess, 0, len(pm.processes))
	for _, p := range pm.processes {

		if p.Cmd.ProcessState != nil && p.Cmd.ProcessState.Exited() {
			continue
		}
		res = append(res, p)
	}
	return res
}

// [EN] Stop terminates a background process by its ID.
// [ID] Stop menghentikan proses latar belakang berdasarkan ID-nya.
func (pm *ProcessManager) Stop(id int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, ok := pm.processes[id]
	if !ok {
		return os.ErrNotExist
	}

	if p.Cmd.Process != nil {
		p.Cmd.Process.Kill()
	}
	delete(pm.processes, id)
	return nil
}

func (pm *ProcessManager) Cleanup() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, p := range pm.processes {
		if p.Cmd.Process != nil {
			p.Cmd.Process.Kill()
		}
	}
	pm.processes = make(map[int]*BackgroundProcess)
}
