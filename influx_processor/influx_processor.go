package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	client "github.com/influxdata/influxdb/client/v2"
)

const (
	INFLUXURL = "http://localhost:8086"
	MYDB      = "prometheus"
	USERNAME  = "root"
	PASSWORD  = "root"
)

type SummaryStats struct {
	mean            float64
	max             float64
	std             float64
	peakToMeanRatio float64
}

// queryDB convenience function to query the database
func queryDB(clnt client.Client, db string, cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: db,
	}
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

func GetSummaryStatsForMetric(clnt client.Client, db string, metric string, tags map[string]string) (*SummaryStats, error) {
	filter := "WHERE "
	for k, v := range tags {
		if len(filter) == 6 { // first tag
			filter = filter + fmt.Sprintf("%s='%s'", k, v)
		} else {
			filter = filter + fmt.Sprintf(" AND %s='%s'", k, v)
		}
	}

	queryCmd := fmt.Sprintf("SELECT mean(*),max(*),stddev(*) from %s %s", metric, filter)
	//fmt.Println("Running query: " + queryCmd)
	result, err := queryDB(clnt, db, queryCmd)
	if err != nil || len(result[0].Series) == 0 {
		return nil, errors.New("Unable to select stats for metric " + metric)
	}
	meanVal, _ := result[0].Series[0].Values[0][1].(json.Number).Float64()
	maxVal, _ := result[0].Series[0].Values[0][2].(json.Number).Float64()
	stdVal, _ := result[0].Series[0].Values[0][3].(json.Number).Float64()

	ss := &SummaryStats{
		mean:            meanVal,
		max:             maxVal,
		std:             stdVal,
		peakToMeanRatio: maxVal / meanVal,
	}

	return ss, nil
}

func main() {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     INFLUXURL,
		Username: USERNAME,
		Password: PASSWORD,
	})
	if err != nil {
		fmt.Println("Unable to create influx http client!")
		log.Fatal(err)
	}

	nodepools := []string{"action-classify", "action-gke", "db", "db-preempt", "druid-preempt", "druid-ssd-preempt", "mixed", "mixed-preempt", "nginx", "ping-gke"}
	metrics := []string{"container_memory_working_set_bytes"}
	// metrics := []string{"container_cpu_usage_seconds_total", "container_memory_working_set_bytes", "container_network_receive_bytes_total", "container_network_transmit_bytes_total", "container_fs_reads_bytes_total", "container_fs_writes_bytes_total"}

	summaryStatsMap := make(map[string]map[string]SummaryStats)
	for _, metric := range metrics {
		summaryStatsMap[metric] = make(map[string]SummaryStats)
		tags := map[string]string{}
		fmt.Printf("Summary stats for metric %s:\n", metric)
		for _, nodepool := range nodepools {
			tags["node_pool"] = nodepool
			summaryStats, err := GetSummaryStatsForMetric(c, MYDB, metric, tags)
			if err != nil {
				fmt.Println("Unable to get summary stats for node_pool " + nodepool + ": " + err.Error())
			} else {
				summaryStatsMap[metric][nodepool] = *summaryStats
				fmt.Printf(" node_pool %s: %+v \n", nodepool, *summaryStats)
			}
		}

	}
}
