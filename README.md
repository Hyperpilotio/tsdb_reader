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


I pushed a new update for tsdb_reader to support filtering metrics based on labels
for example: ./tsdb_reader get_metric_example ~/src/smyte-data/smyte process_cpu_seconds_total 2 2 kubernetes_io_hostname=gke-primary-action-gke-20170808-8ea32133-02df
Labels for series:  {__name__="process_cpu_seconds_total",action_gke="true",beta_kubernetes_io_arch="amd64",beta_kubernetes_io_fluentd_ds_ready="true",beta_kubernetes_io_instance_type="custom-32-65536",beta_kubernetes_io_os="linux",classify="true",cloud_google_com_gke_nodepool="action-gke-20170808",cloud_google_com_gke_preemptible="true",failure_domain_beta_kubernetes_io_region="us-central1",failure_domain_beta_kubernetes_io_zone="us-central1-b",instance="gke-primary-action-gke-20170808-8ea32133-02df",job="kubernetes-nodes",kubernetes_io_hostname="gke-primary-action-gke-20170808-8ea32133-02df",local_redis="true",node_pool="action-gke",preemptible="true",process_action="true",process_sqrl="true",pubsub="true",service_account="true",smyteid_gen="true",workload="action-gke"}
Time and value:  1513746953909 0.99
Time and value:  1513747013909 3.27
No more series
get_metric_example <data_dir> <metric_name> <number_of_series> <number_of_points> <filters_csv>
