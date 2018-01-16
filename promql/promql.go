package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql"
	promtsdb "github.com/prometheus/prometheus/storage/tsdb"
	"github.com/prometheus/tsdb"
	"strings"
)

func parseMsString(value string) (*time.Time, error) {
	timeNum, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse time int: " + err.Error())
	}

	t := time.Unix(0, int64(timeNum*int(time.Millisecond)))
	return &t, nil
}


func printMatrixAsCsv(matrix promql.Matrix) {
	firstLine := true
	columns := []string{"time", "value"}
	csvFormat := "%d,%f"
	for _, series := range matrix {
		if firstLine {
			for _, label := range series.Metric {
				columns = append(columns, label.Name)
				csvFormat += ",%s"
			}
			firstLine = false
			fmt.Println(strings.Join(columns, ","))
		}
		for _, point := range series.Points {
			values := []interface{}{point.T, point.V}
			for _, label := range series.Metric {
				values = append(values, label.Value)
			}
			fmt.Printf(csvFormat+"\n", values...)
		}
	}
}

func printVectorAsCsv(vector promql.Vector) {
	firstLine := true
	columns := []string{"time", "value"}
	csvFormat := "%d,%f"
	for _, sample := range vector {
		if firstLine {
			for _, label := range sample.Metric {
				columns = append(columns, label.Name)
				csvFormat += ",%s"
			}
			firstLine = false
			fmt.Println(strings.Join(columns, ","))
		}
		values := []interface{}{sample.T, sample.V}
		for _, label := range sample.Metric {
			values = append(values, label.Value)
		}
		fmt.Printf(csvFormat +"\n", values...)
	}
}


func main() {
	if len(os.Args) < 4 {
		fmt.Println("promsql <tsdb_path> <query> <time> [<end_time> <interval_seconds>]")
		return
	}

	var minBlockDuration model.Duration
	minBlockDuration.Set("2h")
	w := log.NewSyncWriter(os.Stdout)
	logger := log.NewLogfmtLogger(w)
	options := promql.EngineOptions{
		MaxConcurrentQueries: 10,
		Timeout:              2 * time.Hour,
		Logger:               logger,
	}
	tsdbOptions := tsdb.Options{
		WALFlushInterval:  0,
		RetentionDuration: 0,
		BlockRanges:       tsdb.ExponentialBlockRanges(int64(2*time.Hour)/1e6, 3, 5),
		NoLockfile:        true,
	}
	tsdbPath := os.Args[1]
	fmt.Println(tsdbPath)
	db, err := tsdb.Open(tsdbPath, nil, nil, &tsdbOptions)
	if err != nil {
		fmt.Println("Unable to open tsdb: " + err.Error())
		return
	}
	adapter := promtsdb.Adapter(db, 0)
	engine := promql.NewEngine(adapter, &options)
	queryString := os.Args[2]
	startTime, err := parseMsString(os.Args[3])
	if err != nil {
		fmt.Println("Unable to parse start time: " + err.Error())
		return
	}

	var query promql.Query
	if len(os.Args) >= 5 {
		endTime, err := parseMsString(os.Args[4])
		if err != nil {
			fmt.Println("Unable to parse end time: " + err.Error())
			return
		}
		interval := 5*time.Second
		if len(os.Args) >= 6 {
			intervalSeconds, err := strconv.Atoi(os.Args[5])
			if err != nil {
				fmt.Println("Unable to parse interval seconds: " + err.Error())
				return
			}

			interval = time.Duration(intervalSeconds) * time.Second
		}

		fmt.Println("Running range query with query, start and end time: ", queryString, *startTime, *endTime)
		query, err = engine.NewRangeQuery(queryString, *startTime, *endTime, interval)
	} else {
		query, err = engine.NewInstantQuery(queryString, *startTime)
	}

	if err != nil {
		fmt.Println("Unable to evaluate query: " + err.Error())
		return
	}

	result := query.Exec(context.Background())
	if result.Err != nil {
		fmt.Println("Query exec error: " + result.Err.Error())
		return
	}

	switch result.Value.Type() {
	case promql.ValueTypeVector:
		vector, _ := result.Vector()
		printVectorAsCsv(vector)
	case promql.ValueTypeMatrix:
		matrix, _ := result.Matrix()
		printMatrixAsCsv(matrix)
	}
}
