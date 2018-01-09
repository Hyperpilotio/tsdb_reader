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

type InfluxClient struct {
	influxUrl string
	db        string
	username  string
	password  string
}

func NewInfluxClient(influxUrl string, db string, username string, password string) *InfluxClient {
	client := &InfluxClient{
		influxUrl: influxUrl,
		db:        db,
		username:  username,
		password:  password,
	}

	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxUrl,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, errors.New("Unable to create influx http client: " + err.Error())
	}

	// Create a new database
	_, err = queryDB(c, fmt.Sprintf("CREATE DATABASE %s", db))
	if err != nil {
		return nil, errors.New("Unable to create database: " + err.Error())
	}

	// Create a default retention policy
	_, err = queryDB(c, fmt.Sprintf("CREATE RETENTION POLICY autogen ON %s DURATION 1w REPLICATION 1 DEFAULT", MYDB))
	if err != nil {
		return nil, errors.New("Unable to create retention policy: " + err.Error())
	}

}

func (client *InfluxClient) WritePoint() {

}

func main() {
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
