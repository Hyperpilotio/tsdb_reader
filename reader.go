package main

import (
	"fmt"
	"math"
	"os"

	"github.com/prometheus/tsdb"
)

func main() {
	path := os.Args[1]
	db, err := tsdb.Open(path, nil, nil, nil)
	if err != nil {
		fmt.Println("Unable to create db: " + err.Error())
		return
	}

	q, err := db.Querier(math.MinInt64, math.MaxInt64)
	if err != nil {
		fmt.Println("Unable to create queries: " + err.Error())
		return
	}
	defer q.Close()

	vals, err := q.LabelValues("__name__")
	if err != nil {
		fmt.Println("Unable to get label values: " + err.Error())
		return
	}

	for _, val := range vals {
		fmt.Println(val)
	}
}
