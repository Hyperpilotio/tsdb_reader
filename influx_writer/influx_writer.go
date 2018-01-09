package main

import (
	"fmt"
	"log"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

const (
	INFLUXURL = "http://localhost:8086"
	MYDB      = "prometheus"
	USERNAME  = "root"
	PASSWORD  = "root"
)

// queryDB convenience function to query the database
func queryDB(clnt client.Client, cmd string) (res []client.Result, err error) {
	q := client.NewQuery(cmd, "", "")
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func main() {
	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     INFLUXURL,
		Username: USERNAME,
		Password: PASSWORD,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a new database
	_, err = queryDB(c, fmt.Sprintf("CREATE DATABASE %s", MYDB))
	if err != nil {
		log.Fatal(err)
	}

	// Create a default retention policy
	_, err = queryDB(c, fmt.Sprintf("CREATE RETENTION POLICY autogen ON %s DURATION 1w REPLICATION 1 DEFAULT", MYDB))
	if err != nil {
		log.Fatal(err)
	}

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MYDB,
		Precision: "us",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a point and add to batch
	tags := map[string]string{"cpu": "cpu-total"}
	fields := map[string]interface{}{
		"idle":   10.1,
		"system": 53.3,
		"user":   46.6,
	}

	pt, err := client.NewPoint("cpu_usage", tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := c.Write(bp); err != nil {
		log.Fatal(err)
	}
}
