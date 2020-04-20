package main

import (
	"time"
	"math/rand"
	"os"
	"flag"
	"sync"
	"fmt"
	"strconv"
	
	"collector/monitor"
	"collector/sender"
	"collector/types"
)

/*
	The core of the metric collection agent. The metric collection agent has two components:
	1) a metric collector and 2) a sender. The metric collector reads process metrics from the
	local host and it pushes them on a channel. The sender reads the metrics from the channel
	and sends them to a server. 
	
	The agent also has a simulation mode, in which it generates fake metrics from multiple processes.
	The fake collection agents do not run forever, a parameter configures how many metrics they should
	generate, after which they die. In the simulation, some agents generate a high load of CPU (with a prob
	of 20%). Also, in general agents could have some spikes in utilization.
	
	The metric collector and the sender are pluggable. For the metric collector part, 
	I have written a plugin to read cpu utilization of a
	process from my macos, and another plugin to generate fake metrics. For the sender part,
	I have used an InfluxDB client.
	
	I could have used the Telegraf agent for InfluxDB but I wanted to simulate agents and it seemed easier
	to write everything from sratch.
*/

func runSimulation(sender sender.Sender, myname string, requests chan *types.LocalLoad, 
				maxPollInterval int, nProcs int, maxRequests int) {
					
	var wg sync.WaitGroup
	// create chan to receive requests
	
	stopCh := make(chan struct{})
	for i:=0; i<= nProcs; i++ {
		flipCoinProb := rand.Intn(100)
		flipCoin := false // make random with prob 0.2
		if flipCoinProb > 80 {
			flipCoin = true
		}
		config := &types.FakeMonitorConfig{
			MonitorConfig: types.MonitorConfig {
				PollInterval: maxPollInterval,
				Name:		  myname+"_xxxxx_"+strconv.Itoa(i),
			},
			MaxRequests:	  maxRequests,
			HasHighUsage: flipCoin,
		}
		m := monitor.NewFakeMonitor(config, requests)
		
		wg.Add(1)
		go m.Run(&wg, stopCh)
	}
	go sender.Run(stopCh)
	
	//time.Sleep(10*time.Second)
	// wait until all clients are done
    fmt.Printf("Waiting all to be done \n")
    wg.Wait()
    
    // tell sender to finish
    stopCh <- struct{}{}
    time.Sleep(10*time.Second)
}

func runMonitor(sender sender.Sender, myname string, requests chan *types.LocalLoad, 
				maxPollInterval int, pid string) {
					
	stopCh := make(chan struct{})
	config := &types.MonitorConfig{
				PollInterval: maxPollInterval,
				Name:		  myname,
				ProcessName:  pid,
		}
	
	m := monitor.NewMacOSMonitor(config, requests)
	go m.Run(nil, stopCh)
	go sender.Run(stopCh)
	
	// wait for ever
	<- stopCh
}

func main() {
	
	simulate := flag.NewFlagSet("simulate", flag.ExitOnError) 
    run := flag.NewFlagSet("run", flag.ExitOnError)
    
    maxpollinterval := simulate.Int("maxpoll", 10, "Max poll interval.")
    nprocs 			:= simulate.Int("maxproc", 10, "Maximum processes to simulate.")
    maxrequests		:= simulate.Int("maxreq", 100000, "Number of readings to generate.")
    localaddressim   := simulate.String("host", "localhost", "How do I identify this client on the network.")
    sendersim		:= simulate.String("sendertype", "influxdb", "Type of sender module to use.")
    senderaddr		:= simulate.String("server", "localhost:8080", "Server that collects the metrics.")
    senderconf		:= simulate.String("conf", "default.conf", "Configuration for sender (auth, etc.).")
    
    maxpollintervalr := run.Int("maxpoll", 10, "Max poll interval.")
    procname			:= run.String("procid", "129292", "Process PID to monitor.")
    localaddress		:= run.String("host", "localhost", "How do I identify this client on the network.")
    senderr			:= run.String("sendertype", "influxdb", "Type of sender module to use.")
    senderaddrr		:= run.String("server", "localhost:8080", "Server that collects the metrics.")
    senderconfr		:= run.String("conf", "default.conf", "Configuration for sender (auth, etc.).")
    
    if len(os.Args) < 2 {
        fmt.Println("simulate|run subcommand is required.")
        os.Exit(1)
    }
	command := os.Args[1]
	
	rand.Seed(time.Now().UnixNano())
	
	requests := make(chan *types.LocalLoad)
	
	if command == "simulate" {
		simulate.Parse(os.Args[2:])
		if *maxpollinterval < 2 || *nprocs < 1 || *maxrequests < 1 {
			fmt.Println("Arguments out of range.")
			os.Exit(1)
		}
		if len(*sendersim) == 0 || len(*senderaddr) == 0 {
			fmt.Println("Arguments out of range.")
			os.Exit(1)
		}
		sender := sender.NewGenericSender(*sendersim, &types.SenderConfig{
						Address: *senderaddr,
						ConfFile: *senderconf,
					}, requests)
		if sender == nil  {
			fmt.Printf("Cannot create sender from arguments %s %s\n.", sendersim, senderaddr)
			os.Exit(1)
		}
		runSimulation(*sender, *localaddressim, requests, *maxpollinterval, *nprocs, *maxrequests)
	}
	
	if command == "run" {
		run.Parse(os.Args[2:])
		if *maxpollintervalr < 2 || len(*procname) < 1 || len(*localaddress) < 1{
			fmt.Println("Arguments out of range.")
		}
		if len(*sendersim) == 0 || len(*senderaddr) == 0 {
			fmt.Println("Arguments out of range.")
			os.Exit(1)
		}
		sender := sender.NewGenericSender(*senderr, &types.SenderConfig{
						Address: *senderaddrr,
						ConfFile: *senderconfr,
					}, requests)
		if sender == nil  {
			fmt.Printf("Cannot create sender from arguments %s %s\n.", senderr, senderaddr)
			os.Exit(1)
		}
		runMonitor(*sender, *localaddress, requests, *maxpollintervalr, *procname)
	}
	
	fmt.Println("Command not supported.")
	
}
