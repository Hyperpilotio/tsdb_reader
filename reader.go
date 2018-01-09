package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"
)

func PrintLabelValues(db *tsdb.DB, labelName string) error {
	q, err := db.Querier(math.MinInt64, math.MaxInt64)
	if err != nil {
		return fmt.Errorf("Unable to create queries: " + err.Error())
	}
	defer q.Close()

	vals, err := q.LabelValues(labelName)
	if err != nil {
		return fmt.Errorf("Unable to get label values: " + err.Error())
	}

	for _, val := range vals {
		// We don't need solr data and there are tons of them.
		if !strings.HasPrefix(val, "solr") {
			fmt.Println(val)
		}
	}

	return nil
}

func PrintSeries(db *tsdb.DB) error {
	q, err := db.Querier(math.MinInt64, math.MaxInt64)
	if err != nil {
		return fmt.Errorf("Unable to create queries: " + err.Error())
	}
	defer q.Close()

	set, err := q.Select(labels.NewEqualMatcher("node", "gke-primary-action-classify-uc1b-2017-72cfea2c-6r5b"))
	if err != nil {
		return fmt.Errorf("Unable to select: " + err.Error())
	}

	for set.Next() {
		series := set.At()
		fmt.Println("All labels: %+v", series.Labels())
		iter := series.Iterator()
		iter.Next()
		t, v := iter.At()
		fmt.Println("First point t %d:%f", t, v)
	}

	return nil
}

func main() {
	path := os.Args[1]
	db, err := tsdb.Open(path, nil, nil, nil)
	if err != nil {
		fmt.Println("Unable to create db: " + err.Error())
		return
	}

	/*
		labelName := "__name__"
		if len(os.Args) >= 3 {
			labelName = os.Args[2]
		}
	*/
	//PrintLabelValues(db, labelName)
	PrintSeries(db)
}
