package monitor

import (
	"fmt"
	"os/exec"
	"strings"
	"strconv"
	"bytes"
	"math/rand"
	"time"
	"sync"
	"collector/types"
)
/*
 A basic way to read utilization of a process: I use pgrep and top
 to get cpu "top -pid 53299 -l 1 -stats cpu | tail -1"
 to get pid "pgrep name"
 */

type MacOSMonitor struct {
	config *types.MonitorConfig
	minPollInterval int
	queue chan<- *types.LocalLoad
}

func NewMacOSMonitor(config *types.MonitorConfig,  c chan<- *types.LocalLoad) *MacOSMonitor {
	return &MacOSMonitor {
		config:			config,
		minPollInterval: 1,
		queue: 			 c,
	}
}

func (m *MacOSMonitor) waitRandom() {
	timeToSleep := m.minPollInterval + rand.Intn(m.config.PollInterval - m.minPollInterval)
	time.Sleep(time.Duration(timeToSleep) *  time.Second)
}

func (m *MacOSMonitor) ReadLoad(process string) *types.LocalLoad {
	// get pid for the process
	cmd := exec.Command("pgrep", process)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running pgrep command", err)
		return nil
	}
	pid := strings.Split(out.String(), "\n")
	
	// get cpu for the process top -pid 53299 -l 1 -stats cpu | tail -1
	cmd = exec.Command("top", "-pid", pid[0], "-l", "1", "-stats", "cpu")
	
	var out2 bytes.Buffer
	cmd.Stdout = &out2
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running top command", err)
		return nil
	}
	lines := strings.Split(out2.String(), "\n")
	cpuval, errv := strconv.ParseFloat(strings.Trim(lines[len(lines)-2], " "), 64)
	if errv != nil {
		fmt.Println(errv)
		return nil
	}
	
	return &types.LocalLoad {
		CPU: cpuval,
		Sender: m.config.Name,
	}
	
}

func (m *MacOSMonitor) Run(wg *sync.WaitGroup, stopCh <-chan struct{}) {
	for {
		select {
			case _ = <- stopCh:
				break
			default:
				m.waitRandom()
				currentload := m.ReadLoad(m.config.ProcessName)
				if currentload == nil {
					continue
				}
				m.queue <- currentload
		}
	}
}