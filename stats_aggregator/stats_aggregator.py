import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

INPUT_DIR = "~/data/query-result/"
OUTPUT_DIR = "~/data/summary-stats/"
#RES_LIST = ['cpu', 'mem', 'net_send', 'net_receive', 'disk_read', 'disk_write']
RES_LIST = ['cpu', 'mem']
METRIC_LIST = ['_util_per_instance_95p', '_util_per_instance_max', '_util_per_pool', '_util_per_pod']
COST_MAP = {'action-classify': 0.248, 'action-gke': 1.22, 'db': 0.663, 'db-preempt': 0.663, 'druid-preempt': 0.663,
            'druid-ssd-preempt': 0.704, 'mixed': 0.248, 'mixed-preempt': 0.248, 'nginx': 0.266, 'ping-gke': 0.69}
PERCENTILES = [.5, .95, .99]
#END_TIME = 1514995200917
END_TIME = 1515028900917

class StatsAggregator(object):
    def __init__(self, metric_name):
        self.metric_name = metric_name

    def get_csv_list(self, res_list, data_dir):
        csv_list = {}

        for res in res_list:
            csv_file = data_dir + res + self.metric_name + ".csv"
            csv_list[res] = csv_file

        print("Constructed list of csv filess:", csv_list)
        return csv_list

    def process_csv(self, res, csvfile, out_dir):
        df = pd.read_csv(csvfile, sep=',')
        summary_df = pd.DataFrame()

        for nodepool in df['node_pool'].unique():
            stats_pool = df.loc[(df['node_pool'] == nodepool) & (df['time'] <= END_TIME)]
            summary_df[nodepool] = stats_pool.value.describe(PERCENTILES)
            print("Summarizing %d data points for resource %s, node pool %s"
                  %(len(stats_pool), res, nodepool))

            fig_name = res + self.metric_name + "_" + nodepool
            stats_pool.loc[:, 'time'] = pd.to_datetime(stats_pool['time'], unit='ms')
            stats_pool.plot(x='time', y='value', title=fig_name)
            plt.ylabel('Percent (%)')
            plt.legend().set_visible(False)
            plt.savefig(fig_name+".png")

        outfile = out_dir + res + self.metric_name + ".csv"
        print("\nWriting summary stats of %s resource for all node pools to %s\n" %(res, outfile))
        summary_df.to_csv(outfile)

        plt.close('all')

    def compute_waste(self, res, csvfile1, csvfile2, out_dir):
        df_util = pd.read_csv(csvfile1, sep=',')
        df_num = pd.read_csv(csvfile2, sep=',')
        waste_list = []

        for nodepool in df_util['node_pool'].unique():
            util_pool = df_util.loc[(df_util['node_pool'] == nodepool) & (df_util['time'] <= END_TIME)][['time', 'value']]
            num_pool = df_num.loc[(df_num['node_pool'] == nodepool) & (df_num['time'] <= END_TIME)][['time', 'value']]
            num_avg = num_pool.value.mean()
            print("Average provisioned instances for nodepool %s: %.1f" %(nodepool, num_avg))

            util_pool['time'] = (util_pool['time'] / 1000).astype('int64')
            num_pool['time'] = (num_pool['time'] / 1000).astype('int64')
            df_joined = util_pool.set_index('time').join(num_pool.set_index('time'), how='inner',
                                                         lsuffix='_util', rsuffix='_num')
            waste_num = ( (1 - df_joined.value_util/100) * df_joined.value_num ).mean()
            waste_cost = waste_num * COST_MAP[nodepool]
            waste_list.append({'node pool': nodepool, 'live instances': num_avg,
                               'unused instances': waste_num, 'wasted cost': waste_cost})
            print("Average hourly cost wasted for %s resource in nodepool %s: %.2f" %(res, nodepool, waste_cost))

        outfile = out_dir + res + "_waste_cost.csv"
        waste_df = pd.DataFrame(waste_list)
        waste_df.to_csv(outfile)


if __name__ == "__main__":
    for metric_name in METRIC_LIST:
        aggregator = StatsAggregator(metric_name)
        csv_list = aggregator.get_csv_list(RES_LIST, INPUT_DIR)
        for k, v in csv_list.items():
            aggregator.process_csv(k, v, OUTPUT_DIR)

    aggregator = StatsAggregator("_util_per_pool")
    csv_list = aggregator.get_csv_list(RES_LIST, INPUT_DIR)
    csv2 = INPUT_DIR + "num_instances_per_pool.csv"
    for res, csv1 in csv_list.items():
        aggregator.compute_waste(res, csv1, csv2, OUTPUT_DIR)
