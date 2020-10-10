package mem_index

import (
	"fmt"
	"time"
	"context"
	"reflect"
	"strconv"
	"encoding/json"

	"github.com/jinzhu/gorm"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"golang.org/x/sync/errgroup"
	"github.com/spf13/viper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/deckarep/golang-set"

	"stree-index/pkg"
	"stree-index/pkg/common"
	"stree-index/pkg/config"
	"stree-index/pkg/inverted-index/labels"
	ii "stree-index/pkg/inverted-index"
	"stree-index/pkg/web/controller/node-path"
	"github.com/go-redis/redis"
	"strings"
)

var (
	EcsIdx *ii.HeadIndexReader
	ElbIdx *ii.HeadIndexReader
	RdsIdx *ii.HeadIndexReader
	DcsIdx *ii.HeadIndexReader

	SupportIndex = map[string]bool{
		"ecs": true,
		"elb": true,
		"rds": true,
		"dcs": true,
	}
)

func InitIdx() {

	EcsIdx = ii.NewHeadReader()
	ElbIdx = ii.NewHeadReader()
	RdsIdx = ii.NewHeadReader()
	DcsIdx = ii.NewHeadReader()

}

func FlushAllIdx(logger log.Logger) error {
	start := time.Now()

	level.Info(logger).Log("msg", "FlushAllIdx Start ....")
	ctx, _ := context.WithCancel(context.Background())
	group, _ := errgroup.WithContext(ctx)

	// ecs
	group.Go(func() error {
		err := FlushEcsIdxAdd(logger, []uint64{})
		if err != nil {
			level.Error(logger).Log("msg", "FlushAllIdx FlushEcsIdxAdd Failed", "error", err)
		} else {
			level.Info(logger).Log("msg", "FlushAllIdx FlushEcsIdxAdd Successfully")
		}
		timeTook := time.Since(start)
		pkg.IndexFlushDuration.With(prometheus.Labels{"type": "all", "resource_type": "ecs"}).Set(float64(timeTook.Seconds()))
		level.Info(logger).Log("msg", "FlushEcsIdxAdd", "type", "ecs", "time_took", timeTook)
		return err

	})
	// elb
	group.Go(func() error {

		err := FlushElbIdxAdd(logger, []uint64{})
		if err != nil {
			level.Error(logger).Log("msg", "FlushAllIdx FlushElbIdxAdd Failed", "error", err)
		} else {
			level.Info(logger).Log("msg", "FlushAllIdx FlushElbIdxAdd Successfully")
		}
		timeTook := time.Since(start)
		pkg.IndexFlushDuration.With(prometheus.Labels{"type": "all", "resource_type": "elb"}).Set(float64(timeTook.Seconds()))
		level.Info(logger).Log("msg", "FlushElbIdxAdd", "time_took", timeTook)
		return err

	})

	// rds
	group.Go(func() error {

		err := FlushRdsIdxAdd(logger, []uint64{})
		if err != nil {
			level.Error(logger).Log("msg", "FlushAllIdx FlushRdsIdxAdd Failed", "error", err)
		} else {
			level.Info(logger).Log("msg", "FlushAllIdx FlushRdsIdxAdd Successfully")
		}
		timeTook := time.Since(start)
		pkg.IndexFlushDuration.With(prometheus.Labels{"type": "all", "resource_type": "rds"}).Set(float64(timeTook.Seconds()))
		level.Info(logger).Log("msg", "FlushRdsIdxAdd", "time_took", timeTook)
		return err

	})

	// dcs
	group.Go(func() error {

		err := FlushDcsIdxAdd(logger, []uint64{})
		if err != nil {
			level.Error(logger).Log("msg", "FlushAllIdx FlushDcsIdxAdd Failed", "error", err)
		} else {
			level.Info(logger).Log("msg", "FlushAllIdx FlushDcsIdxAdd Successfully")
		}
		timeTook := time.Since(start)
		pkg.IndexFlushDuration.With(prometheus.Labels{"type": "all", "resource_type": "dcs"}).Set(float64(timeTook.Seconds()))
		level.Info(logger).Log("msg", "FlushDcsIdxAdd", "time_took", timeTook)
		return err

	})

	err := group.Wait()
	timeTook := time.Since(start)
	pkg.IndexFlushDuration.With(prometheus.Labels{"type": "all", "resource_type": "all"}).Set(float64(timeTook.Seconds()))
	level.Info(logger).Log("msg", "FlushAllIndex", "type", "all", "time_took", timeTook)
	return err

}

func mapTolsets(m map[string]string) (labels.Labels) {
	var lset labels.Labels
	for k, v := range m {

		l := labels.Label{
			k,
			v,
		}

		lset = append(lset, l)

	}
	return lset
}
func reflectToLabel(item interface{}) (labels.Labels) {
	t := reflect.TypeOf(item)
	v := reflect.ValueOf(item)

	var lset labels.Labels
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanInterface() {
			//判断是否为可导出字段
			var key, val string
			key = t.Field(i).Tag.Get("gorm")
			fmt.Println(t.Field(i).Type.String())
			switch t.Field(i).Type.String() {
			case "string":

				val = v.Field(i).String()
				fmt.Println(val)
			case "int64":
				fmt.Println(v)
				val = v.Field(i).Interface().(string)
			case "JSON":
				m := make(map[string]string)
				_ = json.Unmarshal([]byte(v.Field(i).Interface().(string)), &m)
				fmt.Println(m)

			}

			l := labels.Label{
				key,
				val,
			}

			lset = append(lset, l)
		}
	}
	return lset
}

func FlushAllIndex(ctx context.Context, logger log.Logger) error {
	ticker := time.NewTicker(time.Duration(viper.GetInt("all_index_update_interval_minute")) * time.Minute)

	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			level.Info(logger).Log("msg", "receive_quit_signal_and_quit")
			return nil
		case <-ticker.C:
			level.Info(logger).Log("msg", "FlushAllIndex Cron Start....")
			FlushAllIdx(logger)
		}

	}
	return nil
}

func UpdateIndex(ctx context.Context, logger log.Logger, cnf *config.RedisKeys) error {
	for {
		select {
		case rType := <-pkg.IndexUpdateChan:
			level.Info(logger).Log("msg", "Receive Sync Signal And Start RunUpdateIndex", "rerouce_type", rType)
			switch rType {
			case "ecs":
				go commonUpdateIndex(cnf.IndexIncrementUpdateKey+rType, rType, FlushEcsIdxAdd, EcsIdx, logger)
			case "elb":
				go commonUpdateIndex(cnf.IndexIncrementUpdateKey+rType, rType, FlushElbIdxAdd, ElbIdx, logger)
			case "rds":
				go commonUpdateIndex(cnf.IndexIncrementUpdateKey+rType, rType, FlushRdsIdxAdd, RdsIdx, logger)
			case "dcs":
				go commonUpdateIndex(cnf.IndexIncrementUpdateKey+rType, rType, FlushDcsIdxAdd, DcsIdx, logger)

			}
		case <-ctx.Done():
			level.Info(logger).Log("msg", "RunReshardHashRingQuit")
			return nil
		}

	}
	return nil
}
func idsToString(ids []uint64) (res string) {
	if len(ids) <= 0 {
		return
	}
	for _, i := range ids {
		res += " " + strconv.FormatUint(i, 10)
	}
	return

}

func commonUpdateIndex(updateRedisKey string, resourceType string, idxAddfunc func(log.Logger, []uint64) error, mmIdx *ii.HeadIndexReader, logger log.Logger) {

	level.Info(logger).Log("msg", "commonUpdateIndexStart...")
	start := time.Now()
	//ctx := context.Background()
	// to_add

	//diffMap := make(map[string][]string)
	var iiU common.IndexIncrementUpdate
	rdb := pkg.GetRedis()
	val, err := rdb.Get(updateRedisKey).Result()
	if err != nil {
		level.Error(logger).Log("msg", "commonUpdateIndexGetDiffIdsFromCacheFailed", "redis_key", updateRedisKey, "resource_type", resourceType, "err", err)
		return
	}

	err = json.Unmarshal([]byte(val), &iiU)
	if err != nil {
		level.Error(logger).Log("msg", "DiffIdsJsonUnmarshalFailed", "resource_type", resourceType, "err", err)
		return
	}
	toAddIds := iiU.ToAdd
	toDelIds := iiU.ToDel
	toModIds := iiU.ToMod
	toDelHashs := iiU.ToDelHash
	pkg.ResouceDiffNumCount.With(prometheus.Labels{"resource_type": resourceType, "mode": "to_add"}).Set(float64(len(toAddIds)))
	pkg.ResouceDiffNumCount.With(prometheus.Labels{"resource_type": resourceType, "mode": "to_del"}).Set(float64(len(toDelIds)))
	pkg.ResouceDiffNumCount.With(prometheus.Labels{"resource_type": resourceType, "mode": "to_mod"}).Set(float64(len(toModIds)))

	if len(toAddIds)+len(toDelIds)+len(toModIds) > 0 {
		level.Info(logger).Log("msg", "UpdateIndexSummary", "resource_type", resourceType, "to_add", len(toAddIds), "to_del", len(toDelIds), "to_mod", len(toModIds), "time_took", time.Since(start))

	}

	// to_add
	if len(toAddIds) > 0 {

		idxAddfunc(logger, toAddIds)
		level.Info(logger).Log("msg", "UpdateIndexToadd", "resource_type", resourceType, "num", len(toAddIds), "ids", idsToString(toAddIds), "time_took", time.Since(start))
	}

	// to_del
	if len(toDelIds) > 0 {
		toDelMap := make(map[uint64]struct{})
		toDelMapHash := make(map[string]struct{})

		for _, id := range toDelIds {
			toDelMap[id] = struct{}{}
		}

		for _, hash := range toDelHashs {
			toDelMapHash[hash] = struct{}{}
		}

		mmIdx.DeleteWithIDs(toDelMap, toDelMapHash)
		level.Info(logger).Log("msg", "UpdateIndexToDel", "resource_type", resourceType, "num", len(toDelIds), "ids", idsToString(toDelIds), "time_took", time.Since(start))

	}

	// to_mod
	if len(toModIds) > 0 {
		toDelMap := make(map[uint64]struct{})
		toDelMapHash := make(map[string]struct{})
		for _, id := range toModIds {
			toDelMap[id] = struct{}{}
		}
		for _, hash := range toDelHashs {
			toDelMapHash[hash] = struct{}{}
		}
		mmIdx.DeleteWithIDs(toDelMap, toDelMapHash)
		idxAddfunc(logger, toModIds)
		level.Info(logger).Log("msg", "UpdateIndexToMod", "resource_type", resourceType, "num", len(toModIds), "ids", idsToString(toModIds), "time_took", time.Since(start))

	}
	timeTook := time.Since(start)
	pkg.IndexFlushDuration.With(prometheus.Labels{"type": "diff", "resource_type": resourceType}).Set(float64(timeTook.Seconds()))
	level.Info(logger).Log("msg", "UpdateIndexFinished ...", "resource_type", resourceType, "time_took", timeTook)

}

//func FlushEcsIdxAdd(logger log.Logger, ids []uint64) error {
//	db := pkg.GetDbCon()
//	var dt *gorm.DB
//	var ecsS []*common.Ecs
//	if len(ids) == 0 {
//		dt = db.Table(viper.GetString("table_ecs")).Find(&ecsS)
//
//	} else {
//		dt = db.Table(viper.GetString("table_ecs")).Where(ids).Find(&ecsS)
//		level.Info(logger).Log("msg", "FlushEcsIdxAddIncrementUpdate")
//	}
//
//	if dt.Error != nil {
//
//		level.Error(logger).Log("msg", "FlushEcsIdxGetALlDataError", "error", dt.Error)
//		return dt.Error
//	}
//	if len(ids) == 0 {
//		pkg.ResouceNumCount.With(prometheus.Labels{"resource_type": "ecs"}).Set(float64(dt.RowsAffected))
//	}
//
//	indexHashM := EcsIdx.GetHashMap()
//	setIndex := mapset.NewSet()
//	for hash, _ := range indexHashM {
//		setIndex.Add(hash)
//	}
//	setDb := mapset.NewSet()
//	dbHashM := make(map[string]uint64, 0)
//	toDelM := make(map[uint64]struct{}, 0)
//	toDelHashM := make(map[string]struct{}, 0)
//	gpANameSet := mapset.NewSet()
//	// setdb - setindex = to_add
//	// setindex - setdb = to_del
//
//	for _, item := range ecsS {
//		dbHashM[item.Hash] = item.ID
//		setDb.Add(item.Hash)
//		if _, loaded := indexHashM[item.Hash]; loaded {
//			// 哈希相同，则无需更新
//			continue
//
//		}
//
//		//if strings.HasPrefix(item.Name, "SGT-hawkeye-") && strings.Contains(item.Name, "etcd") {
//		//	fmt.Println("hash_changes", item)
//		//}
//
//		m := make(map[string]string)
//		m["hash"] = item.Hash
//		tags := make(map[string]string)
//		// 列表型 内网ip 公网ip 安全组
//		piIps := make([]string, 0)
//		puIps := make([]string, 0)
//		secGs := make([]string, 0)
//
//		//dbHashM[item.Hash] = item.ID
//
//		// 单个kv
//
//		m["uid"] = item.Uid
//		m["name"] = item.Name
//		m["cloud_provider"] = item.CloudProvider
//		m["owner_id"] = item.OwnerId
//		m["account_id"] = strconv.FormatUint(item.AccountId, 10)
//		m["instance_type"] = item.InstanceType
//		m["charging_mode"] = item.ChargingMode
//		m["region"] = item.Region
//		m["availability_zone"] = item.AvailabilityZone
//		m["vpc_id"] = item.VpcId
//		m["subnet_id"] = item.SubnetId
//		m["status"] = item.Status
//
//		// json 列表型
//		_ = json.Unmarshal([]byte(item.PrivateIp), &piIps)
//		_ = json.Unmarshal([]byte(item.PublicIp), &puIps)
//		_ = json.Unmarshal([]byte(item.SecurityGroups), &secGs)
//
//		// json map型tag key-value
//		_ = json.Unmarshal([]byte(item.Tags), &tags)
//
//		// 统计gpa
//		//
//		var (
//			nameG string
//			nameP string
//			nameA string
//		)
//		nameG = tags[viper.GetString("tree_g_name")]
//		nameP = tags[viper.GetString("tree_p_name")]
//		nameA = tags[viper.GetString("tree_a_name")]
//		if nameG != "" && nameP != "" && nameA != "" {
//			gpANameSet.Add(fmt.Sprintf("%s.%s.%s", nameG, nameP, nameA))
//		}
//
//		EcsIdx.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
//		EcsIdx.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(tags))
//		for _, ip := range piIps {
//			mp := make(map[string]string)
//			mp["private_ip"] = ip
//			EcsIdx.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(mp))
//		}
//		for _, ip := range puIps {
//			mp := make(map[string]string)
//			mp["public_ip"] = ip
//			EcsIdx.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(mp))
//		}
//
//		for _, ip := range secGs {
//			mp := make(map[string]string)
//			mp["security_groups"] = ip
//			EcsIdx.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(mp))
//		}
//
//	}
//	// 计算to_del
//	if len(ids) == 0 {
//		toDelHash := setIndex.Difference(setDb).ToSlice()
//
//		for _, hash := range toDelHash {
//			id := dbHashM[hash.(string)]
//			toDelM[id] = struct{}{}
//			toDelHashM[hash.(string)] = struct{}{}
//
//		}
//		// 当全量更新时才做del
//		if len(toDelM) > 0 {
//			EcsIdx.DeleteWithIDs(toDelM, toDelHashM)
//		}
//	}
//
//	gpANameS := make([]string, 0)
//
//	for x := range gpANameSet.Iter() {
//		gpa := x.(string)
//
//		gpANameS = append(gpANameS, gpa)
//
//	}
//
//	go tryAddGpaS(gpANameS, "ecs", logger)
//
//	level.Info(logger).Log("msg", "allGpAName", "from", "ecs", "num", len(gpANameS))
//
//	return nil
//
//}

func FlushEcsIdxAdd(logger log.Logger, ids []uint64) error {
	db := pkg.GetDbCon()
	var dt *gorm.DB
	var ecsS []*common.Ecs
	if len(ids) == 0 {
		dt = db.Table(viper.GetString("table_ecs")).Find(&ecsS)

	} else {
		dt = db.Table(viper.GetString("table_ecs")).Where(ids).Find(&ecsS)
		idsStr := make([]string, 0)
		for _, id := range ids {
			idsStr = append(idsStr, strconv.FormatUint(id, 10))
		}

		level.Info(logger).Log("msg", "FlushEcsIdxAddIncrementUpdate", "ids", strings.Join(idsStr, " "))
	}

	if dt.Error != nil {

		level.Error(logger).Log("msg", "FlushEcsIdxGetALlDataError", "error", dt.Error)
		return dt.Error
	}
	if len(ids) == 0 {
		pkg.ResouceNumCount.With(prometheus.Labels{"resource_type": "ecs"}).Set(float64(dt.RowsAffected))
	}
	actuallyH := ii.NewHeadReader()
	gpANameSet := mapset.NewSet()
	// setdb - setindex = to_add
	// setindex - setdb = to_del

	if len(ids) != 0 {
		actuallyH = EcsIdx
	}

	for _, item := range ecsS {

		m := make(map[string]string)
		m["hash"] = item.Hash
		tags := make(map[string]string)
		// 列表型 内网ip 公网ip 安全组
		piIps := make([]string, 0)
		puIps := make([]string, 0)
		secGs := make([]string, 0)

		//dbHashM[item.Hash] = item.ID

		// 单个kv

		m["uid"] = item.Uid
		m["name"] = item.Name
		m["cloud_provider"] = item.CloudProvider
		m["owner_id"] = item.OwnerId
		m["account_id"] = strconv.FormatUint(item.AccountId, 10)
		m["instance_type"] = item.InstanceType
		m["charging_mode"] = item.ChargingMode
		m["region"] = item.Region
		m["availability_zone"] = item.AvailabilityZone
		m["vpc_id"] = item.VpcId
		m["subnet_id"] = item.SubnetId
		m["status"] = item.Status

		// json 列表型
		_ = json.Unmarshal([]byte(item.PrivateIp), &piIps)
		_ = json.Unmarshal([]byte(item.PublicIp), &puIps)
		_ = json.Unmarshal([]byte(item.SecurityGroups), &secGs)

		// json map型tag key-value
		_ = json.Unmarshal([]byte(item.Tags), &tags)

		// 统计gpa
		//
		var (
			nameG string
			nameP string
			nameA string
		)
		nameG = tags[viper.GetString("tree_g_name")]
		nameP = tags[viper.GetString("tree_p_name")]
		nameA = tags[viper.GetString("tree_a_name")]
		if nameG != "" && nameP != "" && nameA != "" {
			gpANameSet.Add(fmt.Sprintf("%s.%s.%s", nameG, nameP, nameA))
		}

		actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(tags))
		for _, ip := range piIps {
			mp := make(map[string]string)
			mp["private_ip"] = ip
			actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(mp))
		}
		for _, ip := range puIps {
			mp := make(map[string]string)
			mp["public_ip"] = ip
			actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(mp))
		}

		for _, ip := range secGs {
			mp := make(map[string]string)
			mp["security_groups"] = ip
			actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(mp))
		}

	}
	// 全量更新
	if len(ids) == 0 {
		start := time.Now()
		//level.Info(logger).Log("msg", "FlushEcsIdxAddRestStart")
		EcsIdx.Reset(actuallyH)
		level.Info(logger).Log("msg", "FlushEcsIdxAddResEnd", "time_took_milliseconds", time.Since(start).Milliseconds())
	}

	gpANameS := make([]string, 0)

	for x := range gpANameSet.Iter() {
		gpa := x.(string)

		gpANameS = append(gpANameS, gpa)

	}

	go tryAddGpaS(gpANameS, "ecs", logger)

	level.Info(logger).Log("msg", "allGpAName", "from", "ecs", "num", len(gpANameS))

	return nil

}

func FlushElbIdxAdd(logger log.Logger, ids []uint64) error {
	db := pkg.GetDbCon()
	var dt *gorm.DB
	var elbS []*common.Elb
	if len(ids) == 0 {
		dt = db.Table(viper.GetString("table_elb")).Find(&elbS)

	} else {
		dt = db.Table(viper.GetString("table_elb")).Where(ids).Find(&elbS)

	}

	if dt.Error != nil {

		level.Error(logger).Log("msg", "FlushElbIdxAddGetALlDataError", "error", dt.Error)
		return dt.Error
	}
	if len(ids) == 0 {
		pkg.ResouceNumCount.With(prometheus.Labels{"resource_type": "elb"}).Set(float64(dt.RowsAffected))
	}
	actuallyH := ii.NewHeadReader()
	gpANameSet := mapset.NewSet()
	// setdb - setindex = to_add
	// setindex - setdb = to_del

	if len(ids) != 0 {
		actuallyH = ElbIdx
	}

	for _, item := range elbS {

		m := make(map[string]string)
		m["hash"] = item.Hash
		tags := make(map[string]string)
		// 列表型 内网ip 公网ip 安全组
		//piIps := make([]string, 0)
		//puIps := make([]string, 0)
		backends := make([]string, 0)
		ports := make([]uint64, 0)
		targetGroup := make([]string, 0)

		// 单个kv

		m["uid"] = item.Uid
		m["name"] = item.Name
		m["cloud_provider"] = item.CloudProvider
		m["account_id"] = strconv.FormatUint(item.AccountId, 10)
		m["region"] = item.Region
		m["status"] = item.Status

		// rds独有的
		//m["master_id"] = strconv.FormatUint(item.MasterId, 10)
		m["ip_address"] = item.IpAddress
		m["dns_name"] = item.DnsName
		m["elb_type"] = item.ElbType

		// json 列表型
		_ = json.Unmarshal([]byte(item.Backends), &backends)
		_ = json.Unmarshal([]byte(item.Port), &ports)
		_ = json.Unmarshal([]byte(item.TargetGroup), &targetGroup)

		// json map型tag key-value
		_ = json.Unmarshal([]byte(item.Tags), &tags)

		// 统计gpa
		//
		var (
			nameG string
			nameP string
			nameA string
		)
		nameG = tags[viper.GetString("tree_g_name")]
		nameP = tags[viper.GetString("tree_p_name")]
		nameA = tags[viper.GetString("tree_a_name")]
		if nameG != "" && nameP != "" && nameA != "" {
			gpANameSet.Add(fmt.Sprintf("%s.%s.%s", nameG, nameP, nameA))
		}

		actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(tags))

		for _, i := range backends {
			m := make(map[string]string)
			m["backends"] = i
			actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		}

		for _, i := range targetGroup {
			m := make(map[string]string)
			m["target_group"] = i
			actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		}

		for _, i := range ports {
			m := make(map[string]string)
			m["port"] = strconv.FormatUint(i, 10)
			actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		}
	}

	// 当全量更新时才做del
	if len(ids) == 0 {
		start := time.Now()
		ElbIdx.Reset(actuallyH)
		level.Info(logger).Log("msg", "FlushElbIdxAddResEnd", "time_took_milliseconds", time.Since(start).Milliseconds())

	}

	gpANameS := make([]string, 0)
	for x := range gpANameSet.Iter() {
		gpa := x.(string)

		gpANameS = append(gpANameS, gpa)

	}

	go tryAddGpaS(gpANameS, "elb", logger)

	level.Info(logger).Log("msg", "allGpAName", "from", "elb", "num", len(gpANameS))

	return nil

}

func FlushRdsIdxAdd(logger log.Logger, ids []uint64) error {
	db := pkg.GetDbCon()
	var dt *gorm.DB
	var rdsS []*common.Rds
	if len(ids) == 0 {
		dt = db.Table(viper.GetString("table_rds")).Find(&rdsS)

	} else {
		dt = db.Table(viper.GetString("table_rds")).Where(ids).Find(&rdsS)

	}

	if dt.Error != nil {

		level.Error(logger).Log("msg", "FlushRdsIdxGetALlDataError", "error", dt.Error)
		return dt.Error
	}
	if len(ids) == 0 {
		pkg.ResouceNumCount.With(prometheus.Labels{"resource_type": "rds"}).Set(float64(dt.RowsAffected))
	}
	actuallyH := ii.NewHeadReader()
	gpANameSet := mapset.NewSet()
	// setdb - setindex = to_add
	// setindex - setdb = to_del

	if len(ids) != 0 {
		actuallyH = RdsIdx
	}

	for _, item := range rdsS {

		m := make(map[string]string)
		m["hash"] = item.Hash
		tags := make(map[string]string)
		// 列表型 内网ip 公网ip 安全组
		piIps := make([]string, 0)
		puIps := make([]string, 0)
		secGs := make([]string, 0)

		// 单个kv

		m["uid"] = item.Uid
		m["name"] = item.Name
		m["cloud_provider"] = item.CloudProvider
		m["account_id"] = strconv.FormatUint(item.AccountId, 10)
		m["instance_type"] = item.InstanceType
		m["architecture_type"] = item.ArchitectureType
		m["charging_mode"] = item.ChargingMode
		m["region"] = item.Region
		m["vpc_id"] = item.VpcId
		m["subnet_id"] = item.SubnetId
		m["status"] = item.Status

		// rds独有的
		m["cluster_id"] = strconv.FormatUint(item.ClusterId, 10)
		//m["master_id"] = strconv.FormatUint(item.MasterId, 10)
		m["master_id"] = item.MasterId
		m["port"] = strconv.FormatUint(item.Port, 10)
		m["cluster_name"] = item.ClusterName
		m["engine"] = item.Engine
		m["engine_version"] = item.EngineVersion
		m["resource_id"] = item.ResourceId
		if item.IsWriter {
			m["is_writer"] = "1"
		} else {
			m["is_writer"] = "0"
		}

		// json 列表型
		_ = json.Unmarshal([]byte(item.PrivateIp), &piIps)
		_ = json.Unmarshal([]byte(item.PublicIp), &puIps)
		_ = json.Unmarshal([]byte(item.SecurityGroups), &secGs)

		// json map型tag key-value
		_ = json.Unmarshal([]byte(item.Tags), &tags)

		// 统计gpa
		//
		var (
			nameG string
			nameP string
			nameA string
		)
		nameG = tags[viper.GetString("tree_g_name")]
		nameP = tags[viper.GetString("tree_p_name")]
		nameA = tags[viper.GetString("tree_a_name")]
		if nameG != "" && nameP != "" && nameA != "" {
			gpANameSet.Add(fmt.Sprintf("%s.%s.%s", nameG, nameP, nameA))
		}

		actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(tags))
		for _, ip := range piIps {
			m := make(map[string]string)
			m["private_ip"] = ip
			actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		}
		for _, ip := range puIps {
			m := make(map[string]string)
			m["public_ip"] = ip
			actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		}

		for _, ip := range secGs {
			m := make(map[string]string)
			m["security_groups"] = ip
			actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		}

	}

	// 当全量更新时才做del
	if len(ids) == 0 {
		// 计算to_del
		start := time.Now()
		RdsIdx.Reset(actuallyH)
		level.Info(logger).Log("msg", "FlushRdsIdxAddResEnd", "time_took_milliseconds", time.Since(start).Milliseconds())

	}

	gpANameS := make([]string, 0)
	for x := range gpANameSet.Iter() {
		gpa := x.(string)

		gpANameS = append(gpANameS, gpa)

	}

	go tryAddGpaS(gpANameS, "rds", logger)

	level.Info(logger).Log("msg", "allGpAName", "from", "rds", "num", len(gpANameS))

	return nil

}

func FlushDcsIdxAdd(logger log.Logger, ids []uint64) error {
	db := pkg.GetDbCon()
	var dt *gorm.DB
	var dcsS []*common.Dcs
	if len(ids) == 0 {
		dt = db.Table(viper.GetString("table_dcs")).Find(&dcsS)

	} else {
		dt = db.Table(viper.GetString("table_dcs")).Where(ids).Find(&dcsS)

	}

	if dt.Error != nil {

		level.Error(logger).Log("msg", "FlushDcsIdxAddGetDataError", "error", dt.Error)
		return dt.Error
	}
	if len(ids) == 0 {
		pkg.ResouceNumCount.With(prometheus.Labels{"resource_type": "dcs"}).Set(float64(dt.RowsAffected))
	}
	actuallyH := ii.NewHeadReader()
	gpANameSet := mapset.NewSet()
	// setdb - setindex = to_add
	// setindex - setdb = to_del

	if len(ids) != 0 {
		actuallyH = DcsIdx
	}

	for _, item := range dcsS {

		m := make(map[string]string)
		m["hash"] = item.Hash
		tags := make(map[string]string)
		// 列表型 内网ip 公网ip 安全组
		piIps := make([]string, 0)
		puIps := make([]string, 0)
		secGs := make([]string, 0)

		// 单个kv

		m["uid"] = item.Uid
		m["name"] = item.Name
		m["cloud_provider"] = item.CloudProvider
		m["account_id"] = strconv.FormatUint(item.AccountId, 10)
		m["instance_type"] = item.InstanceType
		m["charging_mode"] = item.ChargingMode
		m["region"] = item.Region
		m["vpc_id"] = item.VpcId
		m["subnet_id"] = item.SubnetId
		m["status"] = item.Status

		// rds独有的
		m["cluster_id"] = item.ClusterId
		//m["master_id"] = strconv.FormatUint(item.MasterId, 10)
		m["port"] = strconv.FormatUint(item.Port, 10)
		m["engine"] = item.Engine
		m["engine_version"] = item.EngineVersion

		// json 列表型
		_ = json.Unmarshal([]byte(item.PrivateIp), &piIps)
		_ = json.Unmarshal([]byte(item.PublicIp), &puIps)
		_ = json.Unmarshal([]byte(item.SecurityGroups), &secGs)

		// json map型tag key-value
		_ = json.Unmarshal([]byte(item.Tags), &tags)

		// 统计gpa
		//
		var (
			nameG string
			nameP string
			nameA string
		)
		nameG = tags[viper.GetString("tree_g_name")]
		nameP = tags[viper.GetString("tree_p_name")]
		nameA = tags[viper.GetString("tree_a_name")]
		if nameG != "" && nameP != "" && nameA != "" {
			gpANameSet.Add(fmt.Sprintf("%s.%s.%s", nameG, nameP, nameA))
		}

		actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(tags))
		for _, ip := range piIps {
			m := make(map[string]string)
			m["private_ip"] = ip
			actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		}
		for _, ip := range puIps {
			m := make(map[string]string)
			m["public_ip"] = ip
			actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		}

		for _, ip := range secGs {
			m := make(map[string]string)
			m["security_groups"] = ip
			actuallyH.GetOrCreateWithID(item.ID, item.Hash, mapTolsets(m))
		}

	}

	// 当全量更新时才做del
	if len(ids) == 0 {
		// 计算to_del
		start := time.Now()
		DcsIdx.Reset(actuallyH)
		level.Info(logger).Log("msg", "FlushDcsIdxAddResEnd", "time_took_milliseconds", time.Since(start).Milliseconds())

	}

	gpANameS := make([]string, 0)
	for x := range gpANameSet.Iter() {
		gpa := x.(string)

		gpANameS = append(gpANameS, gpa)

	}

	go tryAddGpaS(gpANameS, "dcs", logger)

	level.Info(logger).Log("msg", "allGpAName", "from", "dcs", "num", len(gpANameS))

	return nil

}
func doAddGpa(cacheTargets, targets []string, from string, logger log.Logger) {
	m := make(map[string]struct{})
	for _, i := range cacheTargets {
		m[i] = struct{}{}
	}
	for _, k := range targets {
		if _, loaded := m[k]; !loaded {
			gpa := k
			//level.Info(logger).Log("msg", "gpa_miss_may_be_new", "from", from, "gpa", gpa)
			rc, res := node_path.DbAddNode(common.NodeAddReq{Node: gpa, Level: 4})
			switch rc {
			case 200:
				level.Info(logger).Log("msg", "DbAddNode", "from", from, "gpa", gpa, "rc", rc, "res_str", res)
			case 500:
				level.Error(logger).Log("msg", "DbAddNode", "from", from, "gpa", gpa, "rc", rc, "res_str", res)
			case 400:
				level.Debug(logger).Log("msg", "DbAddNode", "from", from, "gpa", gpa, "rc", rc, "res_str", res)

			}
		}
	}
}

func tryAddGpaS(targets []string, from string, logger log.Logger) {
	level.Info(logger).Log("msg", "tryAddGpaS", "from", from, "num", len(targets))

	cacheTargets := make([]string, 0)
	rdb := pkg.GetRedis()
	val, err := rdb.Get(viper.GetString("all_gpa_cache_key")).Result()
	//if err != nil && err != redis.Nil  {
	//	level.Error(logger).Log("msg", "GetGPAFromCacheFailed", "err", err)
	//	return
	//}
	if err == redis.Nil {
		doAddGpa([]string{}, targets, from, logger)
		return
	}

	err = json.Unmarshal([]byte(val), &cacheTargets)
	if err != nil {
		level.Error(logger).Log("msg", "GetGPAFromCacheUnmarshalFailed", "err", err)
		return
	}

	level.Info(logger).Log("msg", "GetGPAFromCache", "num", len(cacheTargets))
	doAddGpa(cacheTargets, targets, from, logger)
	return
}
