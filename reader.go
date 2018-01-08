package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/prometheus/tsdb"
)

func PrintLabelValues(db *tsdb.DB, metricName string) {
	q, err := db.Querier(math.MinInt64, math.MaxInt64)
	if err != nil {
		fmt.Println("Unable to create queries: " + err.Error())
		return
	}
	defer q.Close()

	vals, err := q.LabelValues(metricName)
	if err != nil {
		fmt.Println("Unable to get label values: " + err.Error())
		return
	}

	for _, val := range vals {
		// We don't need solr data and there are tons of them.
		if !strings.HasPrefix(val, "solr") {
			fmt.Println(val)
		}
	}
}

func main() {
	path := os.Args[1]
	db, err := tsdb.Open(path, nil, nil, nil)
	if err != nil {
		fmt.Println("Unable to create db: " + err.Error())
		return
	}

	labelName := "__name__"
	if len(os.Args) >= 3 {
		labelName = os.Args[2]
	}

	PrintLabelValues(db, labelName)
}
