# tsdb_reader
Prometheus snapshot reader

- Print all values for a label name in all series:
  ./tsdb_reader label_values /Users/tnachen/src/smyte-data/smyte node_pool

- Write data into influx
  ./tsdb_reader write_influx /Users/tnachen/src/smyte-data/smyte