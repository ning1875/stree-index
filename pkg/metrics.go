package pkg

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	IndexFlushDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "index_flush_last_seconds",
		Help: "Duration of index flush  ",
	}, []string{"type", "resource_type"})
	// 全局资源统计
	ResouceNumCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resouce_num_count",
		Help: "Num of resouce",
	}, []string{"resource_type"})

	ResouceRegionNumCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resouce_num_region_count",
		Help: "Num of resouce with region tag",
	}, []string{"resource_type", "region"})

	ResouceCloudProviderNumCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resouce_num_cloud_provider_count",
		Help: "Num of resouce with cloud_provider tag",
	}, []string{"resource_type", "cloud_provider"})

	ResouceDiffNumCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resouce_diff_num_count",
		Help: "Num of resouce",
	}, []string{"resource_type", "mode"})

	GetGPAFromDbDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "get_gpa_from_db_last_duration_seconds",
		Help: "get_gpa_from_db_last_duration_seconds",
	})
	ResourceLastStatisticsDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resource_last_statistics_duration_seconds",
		Help: "resource_last_statistics_duration_seconds",
	}, []string{"type",})

	// gpa 通用资源统计
	GPAAllNumCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpa_all_num_count",
		Help: "Num gpa of all",
	}, []string{"gpa_name", "resource_type"})

	GPAAllNumRegionCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpa_all_region_num_count",
		Help: "Num gpa of all with tag region",
	}, []string{"gpa_name", "resource_type", "region"})
	GPAAllNumCloudProviderCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpa_all_cloud_provider_num_count",
		Help: "Num gpa of all with tag cloud_provider",
	}, []string{"gpa_name", "resource_type", "cloud_provider"})
	GPAAllNumClusterCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpa_all_cluster_num_count",
		Help: "Num gpa of all with tag cluster",
	}, []string{"gpa_name", "resource_type", "cluster"})
	GPAAllNumInstanceTypeCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpa_all_instance_type_num_count",
		Help: "Num gpa of all with tag instance_type",
	}, []string{"gpa_name", "resource_type", "instance_type"})

	GPAEcsNumInstanceTypeCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpa_ecs_instance_type_num_count",
		Help: "Num gpa of all with tag instance_type",
	}, []string{"gpa_name", "cpu_mem", "instance_type"})

	GPAEcsCpuCores = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpa_cpu_cores",
		Help: "Num gpa cpu cores of ecs",
	}, []string{"gpa_name"})

	GPAEcsMemGb = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpa_mem_gbs",
		Help: "Num gpa mem gbs of ecs",
	}, []string{"gpa_name"})
)

func NewMetrics() {
	// last耗时
	prometheus.DefaultRegisterer.MustRegister(IndexFlushDuration)
	prometheus.DefaultRegisterer.MustRegister(GetGPAFromDbDuration)
	prometheus.DefaultRegisterer.MustRegister(ResourceLastStatisticsDuration)
	// 全局资源统计
	prometheus.DefaultRegisterer.MustRegister(ResouceNumCount)
	prometheus.DefaultRegisterer.MustRegister(ResouceRegionNumCount)
	prometheus.DefaultRegisterer.MustRegister(ResouceCloudProviderNumCount)
	prometheus.DefaultRegisterer.MustRegister(ResouceDiffNumCount)

	// gpa 通用资源统计
	prometheus.DefaultRegisterer.MustRegister(GPAAllNumCount)
	prometheus.DefaultRegisterer.MustRegister(GPAAllNumRegionCount)
	prometheus.DefaultRegisterer.MustRegister(GPAAllNumCloudProviderCount)
	prometheus.DefaultRegisterer.MustRegister(GPAAllNumClusterCount)
	prometheus.DefaultRegisterer.MustRegister(GPAAllNumInstanceTypeCount)
	// ecs 特殊
	prometheus.DefaultRegisterer.MustRegister(GPAEcsCpuCores)
	prometheus.DefaultRegisterer.MustRegister(GPAEcsMemGb)
	prometheus.DefaultRegisterer.MustRegister(GPAEcsNumInstanceTypeCount)

	//prometheus.DefaultRegisterer.MustRegister(GPAEcsNumCount)
	//prometheus.DefaultRegisterer.MustRegister(GPAElbNumCount)
	//prometheus.DefaultRegisterer.MustRegister(GPAEcsNumInstanceTypeCount)
	//prometheus.DefaultRegisterer.MustRegister(GPAEcsNumCloudProviderCount)
	//prometheus.DefaultRegisterer.MustRegister(GPAElbNumCloudProviderCount)
	//prometheus.DefaultRegisterer.MustRegister(GPAEcsNumRegionCount)
	//prometheus.DefaultRegisterer.MustRegister(GPAElbNumRegionCount)
	//prometheus.DefaultRegisterer.MustRegister(GPAEcsNumClusterCount)
	//prometheus.DefaultRegisterer.MustRegister(GPAElbNumClusterCount)

}
