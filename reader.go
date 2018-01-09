package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/hyperpilotio/tsdb_reader/influx_writer"
	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"
)

func WriteSeriesToInflux(db *tsdb.DB, prefixes []string) error {
	influxWriter, err := influx_writer.NewInfluxClient("http://localhost:8086", "prometheus", "root", "root")
	if err != nil {
		return errors.New("Unable to create influx client: " + err.Error())
	}

	allMetrics := []string{}
	metricNames, err := GetLabelValues(db, "__name__")
	if err != nil {
		return errors.New("Unable to get all metric names: " + err.Error())
	}

	for _, name := range metricNames {
		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				allMetrics = append(allMetrics, name)
				break
			}
		}
	}

	fmt.Println("Found metrics to write: ", allMetrics)

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
			tags := map[string]string{}
			for _, label := range series.Labels() {
				if label.Name != "__name__" {
					tags[label.Name] = label.Value
				}
			}

			iterator := series.Iterator()
			for iterator.Next() {
				t, v := iterator.At()
				influxWriter.AddBatchPoint(metric, tags, t, v)
			}

			if err := influxWriter.WriteBatch(); err != nil {
				return errors.New("Unable to write batch: " + err.Error())
			}
		}
	}

	return nil
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
	path := os.Args[2]
	options := &tsdb.Options{
		WALFlushInterval:  0,
		RetentionDuration: 0,
		BlockRanges:       tsdb.ExponentialBlockRanges(int64(2*time.Hour)/1e6, 3, 5),
		NoLockfile:        true,
	}

	db, err := tsdb.Open(path, nil, nil, options)
	if err != nil {
		fmt.Println("Unable to create db: " + err.Error())
		return
	}

	action := os.Args[1]

	switch action {
	case "label_values":
		fmt.Println("Printing label values..")
		labelName := "__name__"
		if len(os.Args) >= 4 {
			labelName = os.Args[3]
		}

		labelValues, err := GetLabelValues(db, labelName)
		if err != nil {
			fmt.Println("Unable to get label values: " + err.Error())
			return
		}
		for _, value := range labelValues {
			fmt.Println(value)
		}
		break
	case "write_influx":
		fmt.Println("Writing data into influx..")
		prefixes := []string{"container_", "machine_", "kube_", "net_", "process_"}
		if err := WriteSeriesToInflux(db, prefixes); err != nil {
			fmt.Println("Write data to influx failed: " + err.Error())
			return
		}
	}
}
