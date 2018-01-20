import pandas as pd
import matplotlib.pyplot as plt

INPUT_DIR = "~/data/query-result/"
OUTPUT_DIR = "~/data/summary-stats/"
FIG_DIR = "~/data/summary-figs/"
METRIC_NAME = "_util_per_instance_"
RES_LIST = ['cpu', 'mem', 'net_send', 'net_receive', 'disk_read', 'disk_write']
POOL_LIST = ['action-classify', 'action-gke', 'db', 'db-preempt', 'druid-preempt', 'druid-ssd-preempt',
              'mixed', 'mixed-preempt', 'nginx', 'ping-gke']
STATS_LIST = ['count', 'mean', 'std', 'min', '50%', '90%', '95%', 'max']
PERCENTILES = [.5, .9, .95]

class StatsAggregator(object):
    def __init__(self):
        self.summary_stats = pd.Panel(major_axis=STATS_LIST, minor_axis=POOL_LIST)

    def get_csv_list(self, res_list, data_dir, metric_name, stat_type):
        csv_list = {}

        for res in res_list:
            csv_file = data_dir + res + metric_name + stat_type + ".csv"
            csv_list[res] = csv_file

        print("Constructed list of csv filess:", csv_list)
        return csv_list

    def process_csv(self, res, csvfile, metric_name, stat_type):
        df = pd.read_csv(csvfile, sep=',')
        summary_df = pd.DataFrame()

        for nodepool in df['node_pool'].unique():
            stats_pool = df.loc[df['node_pool'] == nodepool]
            summary_df[nodepool] = stats_pool.value.describe(PERCENTILES)
            print("Summarizing %d data points for resource %s, node pool %s"
                  %(len(stats_pool), res, nodepool))

            fig_name = res + metric_name + stat_type + "_" + nodepool
            stats_pool.loc[:, 'time'] = pd.to_datetime(stats_pool['time'], unit='ms')
            stats_pool.plot(x='time', y='value', title=fig_name)
            plt.ylabel('Percent (%)')
            plt.legend().set_visible(False)
            plt.savefig(fig_name+".png")

        self.summary_stats[res] = summary_df

        outfile = OUTPUT_DIR + res + metric_name + stat_type + ".csv"
        print("\nWriting summary stats of %s resource for all node pools to %s\n" %(res, outfile))
        #self.summary_stats[res].to_csv(outfile)

        plt.close('all')

if __name__ == "__main__":
    aggregator1 = StatsAggregator()
    csv_list1 = aggregator1.get_csv_list(RES_LIST, INPUT_DIR, METRIC_NAME, '95p')
    for k, v in csv_list1.items():
        aggregator1.process_csv(k, v, METRIC_NAME, '95p')

    aggregator2 = StatsAggregator()
    csv_list2 = aggregator2.get_csv_list(RES_LIST, INPUT_DIR, METRIC_NAME, 'max')
    for k, v in csv_list2.items():
        aggregator2.process_csv(k, v, METRIC_NAME, 'max')
