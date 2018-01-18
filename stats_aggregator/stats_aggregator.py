import pandas as pd
import matplotlib.pyplot as plt

OUTPUT_DIR = "~/data/summary-stats/"
CSV_LIST = {'cpu': "~/data/query-result/cpu_util_per_instance_95p.csv",
            'memory': "~/data/query-result/mem_util_per_instance_95p.csv",
            'network_send': "~/data/query-result/net_util_send_per_instance_95p.csv",
            'network_receive': "~/data/query-result/net_util_receive_per_instance_95p.csv",
            'disk_read': "~/data/query-result/disk_util_read_per_instance_95p.csv",
            'disk_write': "~/data/query-result/disk_util_write_per_instance_95p.csv"}
POOL_LIST = ['action-classify', 'action-gke', 'db', 'db-preempt', 'druid-preempt', 'druid-ssd-preempt',
              'mixed', 'mixed-preempt', 'nginx', 'ping-gke']
STATS_LIST = ['count', 'mean', 'std', 'min', '50%', '90%', '95%', 'max']
PERCENTILES = [.5, .9, .95]

class StatsAggregator(object):
    def __init__(self):
        self.summary_stats = pd.Panel(major_axis=STATS_LIST, minor_axis=POOL_LIST)

    def process_csv(self, res, csvfile):
        df = pd.read_csv(csvfile, sep=',')
        summary_df = pd.DataFrame()

        for nodepool in df['node_pool'].unique():
            stats_pool = df.loc[df['node_pool'] == nodepool]
            summary_df[nodepool] = stats_pool.value.describe(PERCENTILES)
            print("Summarizing %d data points for resource %s, node pool %s"
                  %(len(stats_pool), res, nodepool))

        self.summary_stats[res] = summary_df

        outfile = OUTPUT_DIR + "instance_resource_stats_" + res + ".csv"
        print("\nWriting summary stats of %s resource for all node pools to %s\n" %(res, outfile))
        self.summary_stats[res].to_csv(outfile)

if __name__ == "__main__":
    aggregator = StatsAggregator()

    for k, v in CSV_LIST.items():
        aggregator.process_csv(k, v)
