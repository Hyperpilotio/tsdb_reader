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
)

func parseMsString(value string) (*time.Time, error) {
	timeNum, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse time int: " + err.Error())
	}

	t := time.Unix(0, int64(timeNum*int(time.Millisecond)))
	return &t, nil
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("promsql <tsdb_path> <query> <time> [<end_time>]")
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
		query, err = engine.NewRangeQuery(queryString, *startTime, *endTime, 5*time.Second)
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

	fmt.Println("Results:")
	fmt.Println(result.Value)
}
