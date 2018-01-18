package influx_writer

import (
	"errors"
	"fmt"
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
	influxUrl   string
	db          string
	username    string
	password    string
	c           client.Client
	batchPoints client.BatchPoints
}

func NewInfluxClient(influxUrl string, db string, username string, password string) (*InfluxClient, error) {
	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxUrl,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, errors.New("Unable to create influx http client: " + err.Error())
	}

	influxClient := &InfluxClient{
		influxUrl: influxUrl,
		db:        db,
		username:  username,
		password:  password,
		c:         c,
	}

	// Create a new database
	_, err = queryDB(c, fmt.Sprintf("CREATE DATABASE %s", db))
	if err != nil {
		return nil, errors.New("Unable to create database: " + err.Error())
	}

	// Create a default retention policy
	//_, err = queryDB(c, fmt.Sprintf("CREATE RETENTION POLICY autogen ON %s DURATION 4w REPLICATION 1 SHARD DURATION 4w DEFAULT", MYDB))
	//if err != nil {
	//	return nil, errors.New("Unable to create retention policy: " + err.Error())
	//}

	return influxClient, nil
}

func (influxClient *InfluxClient) AddBatchPoint(name string, tags map[string]string, timeMs int64, value float64) error {
	if influxClient.batchPoints == nil {
		// Create a new point batch
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  influxClient.db,
			Precision: "ms",
		})
		if err != nil {
			return errors.New("Unable to create batch points: " + err.Error())
		}

		influxClient.batchPoints = bp
	}

	// Create a point and add to batch
	fields := map[string]interface{}{
		"value": value,
	}

	timestamp := time.Unix(0, timeMs*int64(time.Millisecond))
	pt, err := client.NewPoint(name, tags, fields, timestamp)
	if err != nil {
		return errors.New("Unable to create point: " + err.Error())
	}

	influxClient.batchPoints.AddPoint(pt)

	return nil
}

func (client *InfluxClient) WriteBatch() error {
	if client.batchPoints == nil {
		return errors.New("No batch found")
	}

	if err := client.c.Write(client.batchPoints); err != nil {
		return errors.New("Unable to write batch: " + err.Error())
	}

	client.batchPoints = nil

	return nil
}
