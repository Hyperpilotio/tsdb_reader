# tsdb_reader
Prometheus snapshot reader

- Print all values for a label name in all series:
  ./tsdb_reader label_values /Users/tnachen/src/smyte-data/smyte node_pool

- Write data into influx
  ./tsdb_reader write_influx /Users/tnachen/src/smyte-data/smyte

How to pull smyte data:

- See all the directories:
  gsutil ls gs://hyperpilot-prometheus-backup/smyte

- Download one directory only:
  mkdir 01C1RQ8N3NA5V5KMRBWTAB8V2K/
  cd 01C1RQ8N3NA5V5KMRBWTAB8V2K/
  gsutil rsync gs://hyperpilot-prometheus-backup/smyte/01C1RQ8N3NA5V5KMRBWTAB8V2K/ .
