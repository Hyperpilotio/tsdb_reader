package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hyperpilotio/tsdb_reader/influx_writer"
	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"
)

func GetMetricNames(db *tsdb.DB, prefixes []string) ([]string, error) {
	allMetrics := []string{}
	metricNames, err := GetLabelValues(db, "__name__", nil)
	if err != nil {
		return nil, errors.New("Unable to get all metric names: " + err.Error())
	}

	for _, name := range metricNames {
		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				allMetrics = append(allMetrics, name)
				break
			}
		}
	}

	return allMetrics, nil
}

func PrintAllLabels(db *tsdb.DB, prefixes []string) error {
	allMetrics, err := GetMetricNames(db, prefixes)
	if err != nil {
		return errors.New("Unable to get metric names: " + err.Error())
	}

	for _, metric := range allMetrics {
		set, err := GetSeries(db, map[string]string{"__name__": metric})
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
				tags[label.Name] = label.Value
			}

			jsonString, err := json.Marshal(tags)
			if err != nil {
				return errors.New("Unable to encode json: " + err.Error())
			}

			fmt.Println(string(jsonString))
		}
	}

	return nil
}

func WriteSeriesToInflux(db *tsdb.DB, prefixes []string) error {
	influxWriter, err := influx_writer.NewInfluxClient("http://localhost:8086", "prometheus", "root", "root")
	if err != nil {
		return errors.New("Unable to create influx client: " + err.Error())
	}

	allMetrics, err := GetMetricNames(db, prefixes)
	if err != nil {
		return errors.New("Unable to get metric names: " + err.Error())
	}
	fmt.Println("Found metrics to write: ", allMetrics)

	for _, metric := range allMetrics {
		set, err := GetSeries(db, map[string]string{"__name__": metric})
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

func GetLabelValues(db *tsdb.DB, labelName string, constraint *labels.Label) ([]string, error) {
	q, err := db.Querier(math.MinInt64, math.MaxInt64)
	if err != nil {
		return nil, fmt.Errorf("Unable to create queries: " + err.Error())
	}
	defer q.Close()

	var vals []string

	if constraint != nil {
		vals, err = q.LabelValuesFor(labelName, *constraint)
	} else {
		vals, err = q.LabelValues(labelName)
	}

	if err != nil {
		return nil, fmt.Errorf("Unable to get label values: " + err.Error())
	}

	labelValues := []string{}
	for _, val := range vals {
		labelValues = append(labelValues, val)
	}

	return labelValues, nil
}

func GetSeries(db *tsdb.DB, filter map[string]string) (tsdb.SeriesSet, error) {
	q, err := db.Querier(math.MinInt64, math.MaxInt64)
	if err != nil {
		return nil, fmt.Errorf("Unable to create queries: " + err.Error())
	}
	defer q.Close()

	matchers := []labels.Matcher{}
	for k, v := range filter {
		if strings.Contains(v, "*") {
			matcher, err := labels.NewRegexpMatcher(k, v)
			if err != nil {
				return nil, errors.New("Unable to create regex matcher: " + err.Error())
			}
			matchers = append(matchers, matcher)
		} else {
			matchers = append(matchers, labels.NewEqualMatcher(k, v))
		}

	}

	set, err := q.Select(matchers...)
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
	case "all_labels":
		fmt.Println("Print all labels from all series..")
		prefixes := []string{"container_", "machine_", "kube_", "net_", "process_"}
		if len(os.Args) >= 4 {
			prefixes = strings.Split(os.Args[3], ",")
		}
		PrintAllLabels(db, prefixes)
		break

	case "label_values":
		fmt.Println("Printing label values..")
		labelName := "__name__"
		if len(os.Args) >= 4 {
			labelName = os.Args[3]
		}

		var constraint *labels.Label
		if len(os.Args) >= 5 {
			parts := strings.Split(os.Args[4], "=")
			constraint = &labels.Label{
				Name:  parts[0],
				Value: parts[1],
			}
		}

		labelValues, err := GetLabelValues(db, labelName, constraint)
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
		if len(os.Args) >= 4 {
			prefixes = strings.Split(os.Args[3], ",")
		}
		if err := WriteSeriesToInflux(db, prefixes); err != nil {
			fmt.Println("Write data to influx failed: " + err.Error())
			return
		}
		break
	case "get_metric_example":
		fmt.Println("Get series examples..")
		metricName := os.Args[3]
		seriesCount := 1
		if len(os.Args) >= 5 {
			seriesCount, err = strconv.Atoi(os.Args[4])
			if err != nil {
				fmt.Println("Unable to convert series count to int: " + err.Error())
				return
			}
		}

		metricCount := 1
		if len(os.Args) >= 6 {
			metricCount, err = strconv.Atoi(os.Args[5])
			if err != nil {
				fmt.Println("Unable to convert metric count to int: " + err.Error())
				return
			}
		}

		filter := map[string]string{
			"__name__": metricName,
		}
		if len(os.Args) >= 7 {
			for _, pair := range strings.Split(os.Args[6], ",") {
				parts := strings.Split(pair, "=")
				filter[parts[0]] = parts[1]
			}
		}

		set, err := GetSeries(db, filter)
		if err != nil {
			fmt.Println("Unable to get series: " + err.Error())
			return
		}

		for i := 1; i <= seriesCount; i++ {
			if !set.Next() {
				fmt.Println("No more series")
				return
			}
			series := set.At()
			fmt.Println("Labels for series: ", series.Labels())

			iterator := series.Iterator()
			for j := 1; j <= metricCount; j++ {
				if !iterator.Next() {
					fmt.Println("No more data")
					return
				}
				t, v := iterator.At()
				fmt.Println("Time and value: ", t, v)
			}
		}
		break
	}

}
