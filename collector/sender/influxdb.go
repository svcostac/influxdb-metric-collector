package sender

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"time"
	"collector/types"
	
	"github.com/influxdata/influxdb1-client/v2"
)

type InfluxDBSender struct {
	config *types.SenderConfig
	
	user string
	password string
	database string
	table string
	
	requests <- chan *types.LocalLoad
	//.....
}

func influxDBClient(user string, pass string, address string) client.Client {
    c, err := client.NewHTTPClient(client.HTTPConfig{
        Addr:     address,
        Username: user,
        Password: pass,
    })
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
    return c
}

func parseDBConfigFile(path string) (string, string, string, string, bool) {
	/*
		InfluxDB connection settings are read from a file. If an error is encountered
		the last returned value is false.
	
	*/
	file, err := os.Open(path)
	if err != nil {
	    fmt.Println(err)
	    return "", "", "", "", false
	}
	defer file.Close()
	
	var line string
	reader := bufio.NewReader(file)
	line, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	    return "", "", "", "", false
	}
	s := strings.Split(line, " ")
	if len(s) < 4 {
		fmt.Println("Wrong data in the configuration file.")
		return "", "", "", "", false
	}
	
	user, password, database, table := s[0], s[1], s[2], s[3]
	if len(user) < 1 || len(password) < 1 || len(database) < 1 || len(table) < 1 {
		return "", "", "", "", false
	}
	return user, password, database, table, true
}

func NewInfluxDBSender(config *types.SenderConfig, reqs <- chan *types.LocalLoad) *InfluxDBSender{
	s := &InfluxDBSender{
		config: config,
		requests: reqs,
	}
	// parse config file
	user,pass,db,table,ok := parseDBConfigFile(config.ConfFile)
	if !ok {
		return nil
	}
	
	s.user = user
	s.password = pass
	s.database = db
	s.table = table

	return s	
}

func (s*InfluxDBSender) postRequest(res *types.LocalLoad) {
	c := influxDBClient(s.user, s.password, s.config.Address)
	if c == nil {
		return
	}
	defer c.Close()
	fmt.Printf("%v Sending value %v from %v to Server.\n", time.Now(), res.CPU, res.Sender)
	
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			        Database:  s.database,
			        Precision: "s",
			    })
    if err != nil {
        fmt.Println("Error creating a NewBatchPoint", err)
    }
    
    tags := map[string]string{
            "id":    res.Sender,
		   }
    fields := map[string]interface{}{
            "cpu": res.CPU,
    }
    pt, err := client.NewPoint("cpuload", tags, fields, time.Now())
	if err != nil {
		fmt.Println(err)
	}
    bp.AddPoint(pt)
   
    if err := c.Write(bp); err != nil {
        fmt.Println("Error writing to Influx", err)
    }
	
}

func (s *InfluxDBSender) Run(stopCh <- chan struct{}) {
	// select between requests and done channel, if out of the loop done
	for {
		select {
			case res := <- s.requests:
				go s.postRequest(res)
			case _ = <- stopCh:
				break
		}
	}
}
