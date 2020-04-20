package sender

import (
	"collector/types"
)

type Sender interface {
	
	Run(stopCh <- chan struct{})
	
}

func NewGenericSender(stype string, c *types.SenderConfig, requests <- chan *types.LocalLoad) *Sender {
		var p Sender
		if stype == "influxdb" {
			p = NewInfluxDBSender(c, requests)
			return &p
		}
		return nil
}