#!/bin/bash

promql /var/data/smyte 'max(sum(rate(container_cpu_usage_seconds_total{id="/"}[5m])) by (node_pool, instance) / sum(machine_cpu_cores) by(node_pool, instance) *100) by (node_pool)'  $1 $2 > container_cpu1.csv
promql /var/data/smyte 'max(sum(container_memory_usage_bytes{id="/"}) by (node_pool, instance) / sum(machine_memory_bytes) by(node_pool, instance) *100) by (node_pool)' $1 $2 > container_memory1.csv
promql /var/data/smyte 'max(sum(rate(container_network_transmit_bytes_total[5m])) by (node_pool, instance) / 2e9 * 100) by (node_pool)' $1 $2 > container_network_transmit1.csv
promql /var/data/smyte 'max(sum(rate(container_network_receive_bytes_total[5m])) by (node_pool, instance) / 2e9 * 100) by (node_pool)' $1 $2 > container_network_receive1.csv
promql /var/data/smyte "max(sum(rate(container_fs_reads_bytes_total{node_pool='druid-ssd-preempt'}[5m])) by (node_pool, instance) / 1.56e9 * 100) by (node_pool)"  $1 $2 > container_fs_reads_druid_ssh_preempt1.csv
promql /var/data/smyte "max(sum(rate(container_fs_writes_bytes_total{node_pool='druid-ssd-preempt'}[5m])) by (node_pool, instance)/ 1.56e9 * 100) by (node_pool)" $1 $2 > container_fs_writes_druid_ssh_preempt1.csv
promql /var/data/smyte "max(sum(rate(container_fs_reads_bytes_total{node_pool=~'(db|db-preempt|druid-preempt)'}[5m])) by (node_pool, instance) / 4.8e8 * 100) by (node_=pool)" $1 $2 > container_fs_reads_db_others1.csv
promql /var/data/smyte "max(sum(rate(container_fs_writes_bytes_total{node_pool=~'(db|db-preempt|druid-preempt)'}[5m])) by (node_pool, instance) / 2.4e8 * 100) by (node_pool)" $1 $2 > container_fs_writes_db_others1.csv
promql /var/data/smyte "max(sum(rate(container_fs_reads_bytes_total{node_pool=~'(action-classify|action-gke|mixed|mixed-preempt|nginx|ping-gke)'}[5m])) by (node_pool, instance) / 1.8e8 * 100) by (node_pool)" $1 $2 > container_fs_reads_action_others1.csv
promql /var/data/smyte "max(sum(rate(container_fs_writes_bytes_total{node_pool=~'(action-classify|action-gke|mixed|mixed-preempt|nginx|ping-gke)'}[5m])) by (node_pool, instance) / 1.2e8 * 100) by (node_pool)" $1 $2 > container_fs_writes_action_others1.csv
