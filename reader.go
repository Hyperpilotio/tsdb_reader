package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"
)

func WriteSeriesToInflux(db *tsdb.DB, prefixes []string) error {
	allMetrics := []string{}
	metricNames := GetLabelValues(db, "__name__")
	for _, name := range metricNames {
		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				allMetrics = append(allMetrics, name)
				break
			}
		}
	}

	fmt.Println("Found metrics to write: %+v", allMetrics)

	for _, metric := range allMetrics {
		set, err := GetSeries(db, "__name__", metric)
		if err != nil {
			return errors.New("Unable to get series: " + err.Error())
		}

		for set.Next() {
			if set.Err() != nil {
				return errors.New("Series set error: " + set.Err().Error())
			}

			series := set.At()
			for series.Next() {
				iterator := series.Iterator()
				for iterator.Next() {
					t, v := iterator.At()

				}
			}
		}
	}
}

func GetLabelValues(db *tsdb.DB, labelName string) ([]string, error) {
	q, err := db.Querier(math.MinInt64, math.MaxInt64)
	if err != nil {
		return nil, fmt.Errorf("Unable to create queries: " + err.Error())
	}
	defer q.Close()

	vals, err := q.LabelValues(labelName)
	if err != nil {
		return nil, fmt.Errorf("Unable to get label values: " + err.Error())
	}

	labelValues := []string{}
	for _, val := range vals {
		labelValues = append(labelValues, val)
	}

	return labelValues, nil
}

func GetSeries(db *tsdb.DB, labelName, labelValue string) (tsdb.SeriesSet, error) {
	q, err := db.Querier(math.MinInt64, math.MaxInt64)
	if err != nil {
		return nil, fmt.Errorf("Unable to create queries: " + err.Error())
	}
	defer q.Close()

	//set, err := q.Select(labels.NewEqualMatcher("node", "gke-primary-action-classify-uc1b-2017-72cfea2c-6r5b"))
	set, err := q.Select(labels.NewEqualMatcher(labelName, labelValue))
	if err != nil {
		return nil, fmt.Errorf("Unable to select: " + err.Error())
	}

	return set, nil
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
	//PrintSeries(db)
	prefixes := []string{"container_", "machine_", "kube_", "net_", "process_"}
	WriteSeriesToInflux(db, prefixes)
}
