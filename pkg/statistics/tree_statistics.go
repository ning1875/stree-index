package statistics

import (
	"stree-index/pkg/common"
	"github.com/jinzhu/gorm"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"fmt"
	"time"
	"context"
	"strings"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"stree-index/pkg"
	"stree-index/pkg/mem-index"
	"stree-index/pkg/web/controller/stree"
	"sync"
	"stree-index/pkg/inverted-index"
	"encoding/json"
)

func GetInstanceType(logger log.Logger) (m map[string]common.InstanceType) {
	var dt *gorm.DB
	var res []common.InstanceType
	db := pkg.GetDbCon()
	dt = db.Table(viper.GetString("table_cloud_instance_type")).Where(map[string]interface{}{"resource_type": "ecs"}).Find(&res)
	if dt.Error != nil {

		level.Error(logger).Log("msg", "GetInstanceType", "error", dt.Error)
		return
	}
	m = make(map[string]common.InstanceType)
	for _, i := range res {
		m[i.InstanceType] = i
	}
	level.Info(logger).Log("msg", "GetInstanceTypeFromDb", "num", dt.RowsAffected)
	return m
}

func TreeNodeStatistics(ctx context.Context, logger log.Logger) error {
	ticker := time.NewTicker(time.Duration(viper.GetInt("tree_node_statistics_interval_minute")) * time.Minute)
	insM := GetInstanceType(logger)
	statisticsWork(logger, insM)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			level.Info(logger).Log("msg", "receive_quit_signal_and_quit")
			return nil
		case <-ticker.C:
			level.Info(logger).Log("msg", "statisticsWork Cron Start....")
			statisticsWork(logger, insM)
		}

	}
	return nil
}

func statisticsWork(logger log.Logger, insM map[string]common.InstanceType) {
	go GPAStatistics(logger, insM)
	type ss struct {
		table string
		rtype string
	}

	tables := []ss{
		ss{table: viper.GetString("table_ecs"), rtype: "ecs"},
		ss{table: viper.GetString("table_elb"), rtype: "elb"},
		ss{table: viper.GetString("table_rds"), rtype: "rds"},
		ss{table: viper.GetString("table_dcs"), rtype: "dcs"},
	}
	for _, t := range tables {

		al := allStatistics{
			tableName:           t.table,
			resourceType:        t.rtype,
			metricRegion:        pkg.ResouceRegionNumCount,
			metricCloudProvider: pkg.ResouceCloudProviderNumCount,
		}
		go resourceAllStatistics(al, logger)
	}

}

type allStatistics struct {
	tableName           string
	resourceType        string
	metricRegion        *prometheus.GaugeVec
	metricCloudProvider *prometheus.GaugeVec
}

func resourceAllStatistics(as allStatistics, logger log.Logger) {
	level.Info(logger).Log("msg", "resourceAllStatistics Start....", "table", as.tableName)
	start := time.Now()

	type pro struct {
		CloudProvider string `gorm:"cloud_provider"`
		Count         uint64 `gorm:"count"`
	}
	type rei struct {
		Region string `gorm:"region"`
		Count  uint64 `gorm:"count"`
	}

	var (
		pros []pro
		regs []rei
	)
	db := pkg.GetDbCon()

	// cloud_provider
	dt := db.Table(as.tableName).Select("cloud_provider,count(cloud_provider) as count").Group("cloud_provider").Scan(&pros)
	if dt.Error != nil {

		level.Error(logger).Log("msg", "allStatistics_cloud_provider_tag", "error", dt.Error)
		return
	}

	for _, i := range pros {

		as.metricCloudProvider.With(prometheus.Labels{"resource_type": as.resourceType, "cloud_provider": i.CloudProvider}).Set(float64(i.Count))

	}
	// region
	dt = db.Table(as.tableName).Select("region,count(region) as count").Group("region").Scan(&regs)

	if dt.Error != nil {

		level.Error(logger).Log("msg", "ecsAllStatistics_cloud_region_tag", "error", dt.Error)
		return
	}

	for _, i := range regs {

		as.metricRegion.With(prometheus.Labels{"resource_type": as.resourceType, "region": i.Region}).Set(float64(i.Count))

	}

	timeTook := time.Since(start)
	pkg.ResourceLastStatisticsDuration.With(prometheus.Labels{"type": as.tableName}).Set(float64(timeTook.Seconds()))
	level.Info(logger).Log("msg", "resourceAllStatistics End", "table", as.tableName, "time_took", timeTook.Seconds())

}

func gpaCommonStatistics(gpa string, resourceType string, matcher []*common.SingleTagReq, mmIdx *inverted_index.HeadIndexReader, logger log.Logger, m map[string]common.InstanceType) {

	gInputs := common.QueryReq{
		ResourceType: resourceType,
		Labels:       matcher,
		UseIndex:     true,
	}

	matchIdsG := stree.GetIndex(gInputs)

	if len(matchIdsG) == 0 {
		return
	}

	pkg.GPAAllNumCount.With(prometheus.Labels{"gpa_name": gpa, "resource_type": resourceType}).Set(float64(len(matchIdsG)))
	// region
	statsRe := mmIdx.GetGroupDistributionByLabel("region", matchIdsG)
	for _, x := range statsRe.Group {

		pkg.GPAAllNumRegionCount.With(prometheus.Labels{"gpa_name": gpa, "resource_type": resourceType, "region": x.Name}).Set(float64(x.Value))

	}

	// cloud_provider
	statsCp := mmIdx.GetGroupDistributionByLabel("cloud_provider", matchIdsG)

	for _, x := range statsCp.Group {

		pkg.GPAAllNumCloudProviderCount.With(prometheus.Labels{"gpa_name": gpa, "resource_type": resourceType, "cloud_provider": x.Name}).Set(float64(x.Value))

	}

	// cluster
	statsCl := mmIdx.GetGroupDistributionByLabel("cluster", matchIdsG)

	for _, x := range statsCl.Group {

		pkg.GPAAllNumClusterCount.With(prometheus.Labels{"gpa_name": gpa, "resource_type": resourceType, "cluster": x.Name}).Set(float64(x.Value))

	}

	// instance_type
	statsIn := mmIdx.GetGroupDistributionByLabel("instance_type", matchIdsG)

	switch resourceType {
	case "ecs":
		var cpus uint64
		var memGb uint64
		for _, x := range statsIn.Group {

			// 根据规格确定cpu/内存 数据

			if ins, loaded := m[x.Name]; loaded {
				cpuMemStr := fmt.Sprintf("%dc_%dg", ins.Cpu, ins.Mem)
				pkg.GPAEcsNumInstanceTypeCount.With(prometheus.Labels{"gpa_name": gpa, "instance_type": x.Name, "cpu_mem": cpuMemStr}).Set(float64(x.Value))
				cpus += ins.Cpu * x.Value
				memGb += ins.Mem * x.Value
			} else {
				level.Warn(logger).Log("msg", "load_instance_from_cache_failed....", "instance_type", x.Name, "gpa_name", gpa)
			}
		}
		if cpus > 0 {
			pkg.GPAEcsCpuCores.With(prometheus.Labels{"gpa_name": gpa}).Set(float64(cpus))
		}
		if memGb > 0 {
			pkg.GPAEcsMemGb.With(prometheus.Labels{"gpa_name": gpa}).Set(float64(memGb))
		}

	default:
		for _, x := range statsIn.Group {

			pkg.GPAAllNumInstanceTypeCount.With(prometheus.Labels{"gpa_name": gpa, "resource_type": resourceType, "instance_type": x.Name}).Set(float64(x.Value))

		}

	}

}

func GPAStatistics(logger log.Logger, m map[string]common.InstanceType) {

	level.Info(logger).Log("msg", "GPAStatistics Start....")
	start := time.Now()
	targets := GetGpaSFromDb(logger)

	for _, gpa := range targets {
		ss := strings.Split(gpa, ".")
		if len(ss) != 3 {
			continue
		}
		g := ss[0]
		p := ss[1]
		a := ss[2]
		csG := &common.SingleTagReq{
			Type:  1,
			Key:   viper.GetString("tree_g_name"),
			Value: g,
		}
		csP := &common.SingleTagReq{
			Type:  1,
			Key:   viper.GetString("tree_p_name"),
			Value: p,
		}
		csA := &common.SingleTagReq{
			Type:  1,
			Key:   viper.GetString("tree_a_name"),
			Value: a,
		}
		matcherG := []*common.SingleTagReq{
			csG,
		}
		matcherGP := []*common.SingleTagReq{
			csG,
			csP,
		}
		matcherGPA := []*common.SingleTagReq{
			csG,
			csP,
			csA,
		}
		// g stats
		GPASingleTypeSS(g, matcherG, logger, m)
		// p stats
		GPASingleTypeSS(g+"."+p, matcherGP, logger, m)
		// a stats
		GPASingleTypeSS(g+"."+p+"."+a, matcherGPA, logger, m)

	}
	timeTook := time.Since(start)
	pkg.ResourceLastStatisticsDuration.With(prometheus.Labels{"type": "gpa"}).Set(float64(timeTook.Seconds()))
	level.Info(logger).Log("msg", "GPAStatistics End", "all_gpa_num", len(targets), "time_took", timeTook.Seconds())

}

func GPASingleTypeSS(gpa string, matcher []*common.SingleTagReq, logger log.Logger, m map[string]common.InstanceType) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {

		gpaCommonStatistics(gpa, "ecs", matcher, mem_index.EcsIdx, logger, m)
		//EcsStatistics(gpa, matcher, logger, m)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		gpaCommonStatistics(gpa, "elb", matcher, mem_index.ElbIdx, logger, m)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		gpaCommonStatistics(gpa, "rds", matcher, mem_index.RdsIdx, logger, m)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		gpaCommonStatistics(gpa, "dcs", matcher, mem_index.DcsIdx, logger, m)
		wg.Done()
	}()

	wg.Wait()

}

func GetGpaSFromDb(logger log.Logger) (targets []string) {
	start := time.Now()
	var dt *gorm.DB
	var pTns []common.PathTree
	db := pkg.GetDbCon()
	dt = db.Table(viper.GetString("table_path_tree")).Where("level IN (?)", []uint64{common.NodeLevelMap["g"], common.NodeLevelMap["p"], common.NodeLevelMap["a"]}).Find(&pTns)
	if dt.Error != nil {

		level.Error(logger).Log("msg", "GetgpasError", "error", dt.Error)
		return
	}
	existMapGS := make(map[string]common.PathTree)
	existMapPI := make(map[uint64]common.PathTree)
	existMapAI := make(map[uint64]common.PathTree)

	for _, ptn := range pTns {
		switch ptn.Level {
		case common.NodeLevelMap["g"]:
			existMapGS[ptn.Path] = ptn
		case common.NodeLevelMap["p"]:
			existMapPI[ptn.ID] = ptn
		case common.NodeLevelMap["a"]:
			existMapAI[ptn.ID] = ptn

		}

	}

	for gPath, g := range existMapGS {

		for pid, p := range existMapPI {

			pPath := fmt.Sprintf("%s/%d", gPath, pid)
			if pPath == p.Path {

				for aid, a := range existMapAI {
					aPath := fmt.Sprintf("%s/%d", pPath, aid)
					if aPath == a.Path {
						targets = append(targets, fmt.Sprintf("%s.%s.%s", g.NodeName, p.NodeName, a.NodeName))
					}
				}

			}
		}

	}
	timeTook := time.Since(start)
	pkg.GetGPAFromDbDuration.Set(float64(timeTook.Seconds()))

	if len(targets) == 0 {
		level.Info(logger).Log("msg", "GetZeroGpA")
		return
	}
	// set redis
	rdb := pkg.GetRedis()
	//ctx := context.Background()

	data, _ := json.Marshal(targets)

	err := rdb.Set(viper.GetString("all_gpa_cache_key"), data, 0).Err()
	if err != nil {
		level.Error(logger).Log("msg", "Set all gpa  error", "error", err)
	}

	level.Info(logger).Log("msg", "GetGpaSFromDbRes", "time_took", timeTook, "num", len(targets), "num", len(targets))
	return

}
