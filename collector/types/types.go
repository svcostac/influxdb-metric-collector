package types

type LocalLoad struct {
	Sender string
	CPU float64
}

type SenderConfig struct {
	Address  string  // address of the server to connect to
	ConfFile string // additional configuration info
}

type MonitorConfig struct {
	PollInterval int  // maximum time interval at which a reading should be done
	Name string   // name of this client, for metric collection at the server
	ProcessName string // name of the process to monitor on this host
}

type FakeMonitorConfig struct {
	MonitorConfig
	
	MaxRequests int  // how many fake utilization readings to perform
	HasHighUsage bool // if this fake client should generate high cpu utilization or not
}

