package monitor

import (
	"sync"
	"collector/types"
)

type LoadMonitor interface {
	
	ReadLoad(process string) *types.LocalLoad
	
	Run(wg *sync.WaitGroup, stopCh<- chan struct{})
}