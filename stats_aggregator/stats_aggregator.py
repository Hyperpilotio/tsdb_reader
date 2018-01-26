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
END_TIME = 1514995200917
#END_TIME = 1515028900917

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


    def compute_waste_res(self, res, csv_util, csv_num, out_dir):
        df_util = pd.read_csv(csv_util, sep=',')
        df_num = pd.read_csv(csv_num, sep=',')
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

        outfile = out_dir + res + '_waste_cost' + ".csv"
        waste_df = pd.DataFrame(waste_list)
        waste_df.to_csv(outfile)


    def compute_waste_mixed(self, res_list, csv_list, csv_num, out_dir):
        if len(res_list) > 2:
            print("Cannot combine more than two resources!")
            return

        df_util1 = pd.read_csv(csv_list[res_list[0]], sep=',')
        df_util2 = pd.read_csv(csv_list[res_list[1]], sep=',')
        df_num = pd.read_csv(csv_num, sep=',')
        waste_list = []

        for nodepool in COST_MAP.keys():
            util1_pool = df_util1.loc[(df_util1['node_pool'] == nodepool) & (df_util1['time'] <= END_TIME)][['time', 'value']]
            util2_pool = df_util2.loc[(df_util2['node_pool'] == nodepool) & (df_util2['time'] <= END_TIME)][['time', 'value']]
            util1_pool['time'] = (util1_pool['time'] / 1000).astype('int64')
            util2_pool['time'] = (util2_pool['time'] / 1000).astype('int64')
            df_mixed = util1_pool.set_index('time').join(util2_pool.set_index('time'), how='outer',
                                                         lsuffix=res_list[0], rsuffix=res_list[1]).fillna(0)
            df_mixed['max_util'] = df_mixed[['value'+res_list[0], 'value'+res_list[1]]].max(axis=1)

            num_pool = df_num.loc[(df_num['node_pool'] == nodepool) & (df_num['time'] <= END_TIME)][['time', 'value']]
            num_avg = num_pool.value.mean()
            print("Average provisioned instances for nodepool %s: %.1f" %(nodepool, num_avg))
            num_pool['time'] = (num_pool['time'] / 1000).astype('int64')

            df_joined = df_mixed.join(num_pool.set_index('time'), how='inner',
                                                         lsuffix='_util', rsuffix='_num')
            waste_num = ( (1 - df_joined.max_util/100) * df_joined.value ).mean()
            waste_cost = waste_num * COST_MAP[nodepool]
            waste_list.append({'node pool': nodepool, 'live instances': num_avg,
                               'unused instances': waste_num, 'wasted cost': waste_cost})
            print("Average hourly cost wasted in nodepool %s: %.2f" %(nodepool, waste_cost))

        outfile = out_dir + 'waste_cost_mixed' + ".csv"
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
    csv_num = INPUT_DIR + "num_instances_per_pool.csv"
    for res, csv_res in csv_list.items():
        aggregator.compute_waste_res(res, csv_res, csv2, OUTPUT_DIR)

    aggregator = StatsAggregator("_util_per_pool")
    csv_list = aggregator.get_csv_list(RES_LIST, INPUT_DIR)
    csv_num = INPUT_DIR + "num_instances_per_pool.csv"
    aggregator.compute_waste_mixed(RES_LIST, csv_list, csv_num, OUTPUT_DIR)
