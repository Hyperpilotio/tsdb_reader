package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage/tsdb"
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

	options := promql.EngineOptions{}
	tsdbOptions := tsdb.Options{
		WALFlushInterval: 0,
		Retention:        0,
		NoLockfile:       true,
	}
	storage := &tsdb.ReadyStorage{}
	db, err := tsdb.Open(os.Args[1], nil, nil, &tsdbOptions)
	if err != nil {
		fmt.Println("Unable to open tsdb: " + err.Error())
		return
	}
	storage.Set(db, 0)
	engine := promql.NewEngine(storage, &options)
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
		fmt.Println("Query exec error: " + err.Error())
		return
	}

	fmt.Println("Results:")
	fmt.Println(result.Value)
}
