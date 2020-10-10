package node_path

import (
	"fmt"
	"sort"
	"strings"
	"strconv"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/gin-gonic/gin"

	"stree-index/pkg"
	"stree-index/pkg/common"
	"github.com/spf13/viper"

)

func getMaxId() (id uint64, err error) {
	db := pkg.GetDbCon()
	var dt *gorm.DB
	type mid struct {
		Mid uint64
	}
	var Iid mid
	dt = db.Table(viper.GetString("table_path_tree")).Raw(fmt.Sprintf("select max(id) mid from  %s", viper.GetString("table_path_tree"))).Scan(&Iid)
	if dt.Error != nil {

		return 0, dt.Error
	}
	return Iid.Mid, nil
}

func DbAddNode(inputs common.NodeAddReq) (int, string) {

	if _, found := common.NodeLevelReverseMap[inputs.Level]; !found {
		return http.StatusBadRequest, fmt.Sprintf("level invalid:%d", inputs.Level)

	}
	ss := strings.Split(inputs.Node, ".")
	ll := len(ss)
	if int(inputs.Level) != ll+1 {
		return http.StatusBadRequest, fmt.Sprintf("node does not match  level:%d %s", inputs.Level, inputs.Node)
	}

	var pTns []common.PathTree
	var dt *gorm.DB
	db := pkg.GetDbCon()
	// 先判断g.p.a是否存在

	dt = db.Table(viper.GetString("table_path_tree")).Find(&pTns)
	if dt.Error != nil {

		return http.StatusInternalServerError, dt.Error.Error()
	}

	// 查询a
	existMapA := make(map[string]common.PathTree)

	g := ss[0]
	p := ss[1]
	a := ss[2]
	gExist := false
	pExist := false
	aExist := false

	var existP common.PathTree
	var existG common.PathTree

	for _, ptn := range pTns {
		switch ptn.Level {
		case common.NodeLevelMap["g"]:
			if g == ptn.NodeName {
				existG = ptn
				gExist = true
			}

		case common.NodeLevelMap["p"]:
			if p == ptn.NodeName {
				existP = ptn
				pExist = true
			}
		case common.NodeLevelMap["a"]:

			if a == ptn.NodeName {
				existMapA[ptn.Path] = ptn
			}

		}

	}
	// 判断p是否存在

	if pExist {

		for key, v := range existMapA {

			aPath := fmt.Sprintf("%s/%d", existP.Path, v.ID)
			if aPath == key {
				aExist = true
				break
			}

		}
		// p下有a了，g.p.a存在
		if aExist {
			return http.StatusBadRequest, fmt.Sprintf(
				"path_tree:%s aleadly exists", inputs.Node)
		}
		// p存在但是p下面a不存在，插入a
		// 插入a
		thid, e := getMaxId()
		if e != nil {

			return http.StatusInternalServerError, fmt.Sprintf(
				"p exists a  not add a get maxid  failed:%s", e)
		}
		thid += 1
		path := fmt.Sprintf("%s/%d", existP.Path, thid)
		newA := common.PathTree{
			ID:       thid,
			Level:    common.NodeLevelMap["a"],
			Path:     path,
			NodeName: a,
		}
		// 插入a
		dt = db.Table(viper.GetString("table_path_tree")).Create(&newA)
		if dt.Error != nil {
			return http.StatusInternalServerError, fmt.Sprintf(
				"p exists a  not add a   failed:%s", dt.Error)

		}
		return http.StatusOK, fmt.Sprintf(
			"g p exists add a  successfully:%s", inputs.Node)

	}
	// p不存在，判断g是否存在

	if !gExist {
		// g不存在 插入g
		// 插入g
		thid, e := getMaxId()
		if e != nil {
			return http.StatusInternalServerError, fmt.Sprintf(
				"g not exists get maxid  failed:%s", e)
		}
		thid += 1
		path := fmt.Sprintf("/1/%d", thid)
		newG := common.PathTree{
			ID:       thid,
			Level:    common.NodeLevelMap["g"],
			Path:     path,
			NodeName: g,
		}
		existG = newG
		dt = db.Table(viper.GetString("table_path_tree")).Create(&newG)
		if dt.Error != nil {
			return http.StatusInternalServerError, fmt.Sprintf(
				"g not exists  add g    failed:%s", dt.Error)
		}

	}
	// 插入p 和 a
	// 插入p

	thid, e := getMaxId()
	if e != nil {
		return http.StatusBadRequest, fmt.Sprintf(
			"g p a not exists add p get maxid  failed:%s", e)

	}
	thid += 1
	path := fmt.Sprintf("%s/%d", existG.Path, thid)
	newP := common.PathTree{
		ID:       thid,
		Level:    common.NodeLevelMap["p"],
		Path:     path,
		NodeName: p,
	}

	dt = db.Table(viper.GetString("table_path_tree")).Create(&newP)
	if dt.Error != nil {
		return http.StatusBadRequest, fmt.Sprintf(
			"g p a not exists add p after g   failed:%s", dt.Error)
	}
	existP = newP
	// 插入a
	thid, e = getMaxId()
	if e != nil {
		return http.StatusInternalServerError, fmt.Sprintf(
			"p exists a  not add a get maxid  failed:%s", e)
	}
	thid += 1
	path = fmt.Sprintf("%s/%d", existP.Path, thid)
	newA := common.PathTree{
		ID:       thid,
		Level:    common.NodeLevelMap["a"],
		Path:     path,
		NodeName: a,
	}
	// 插入a
	dt = db.Table(viper.GetString("table_path_tree")).Create(&newA)
	if dt.Error != nil {
		return http.StatusInternalServerError, fmt.Sprintf(
			"g p a not exists add a   failed:%s", dt.Error)

	}
	return http.StatusOK, fmt.Sprintf(
		" p a not exists add a  successfully:%s", inputs.Node)

}

func AddNodeByGetAll(c *gin.Context) {
	var inputs common.NodeAddReq
	if err := c.Bind(&inputs); err != nil {
		common.JSONR(c, http.StatusBadRequest, err)
		return
	}

	code, reStr := DbAddNode(inputs)
	common.JSONR(c, code, reStr)
	return

}

func DeleteNode(c *gin.Context) {
	/*
	node=a                 & level=2
	node=a.cicd            & level=3
	node=a.cicd.jenkins    & level=4

	*/
	var inputs common.NodeAddReq
	if err := c.Bind(&inputs); err != nil {
		common.JSONR(c, http.StatusBadRequest, err)
		return
	}
	// 判断gpa是否合理
	if _, found := common.NodeLevelReverseMap[inputs.Level]; !found {
		common.JSONR(c, http.StatusBadRequest, fmt.Sprintf("level invalid:%d", inputs.Level))
		return
	}
	ss := strings.Split(inputs.Node, ".")
	ll := len(ss)
	if int(inputs.Level) != ll+1 {
		common.JSONR(c, http.StatusBadRequest, fmt.Sprintf("node does not match  level:%d %s", inputs.Level, inputs.Node))
		return
	}

	var dt *gorm.DB
	db := pkg.GetDbCon()
	switch inputs.Level {
	case common.NodeLevelMap["g"]:
		// 说明要删除g，先查询g
		var pathTreeNode common.PathTree
		g := ss[0]
		dt = db.Table(viper.GetString("table_path_tree")).Where(map[string]interface{}{"level": inputs.Level, "node_name": g}).First(&pathTreeNode)
		if dt.Error != nil && !dt.RecordNotFound() {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"query g failed:%s", dt.Error))
			return
		}
		if dt.RowsAffected == 0 {
			common.JSONR(c, http.StatusBadRequest, fmt.Sprintf(
				"g not exists no need to del:%s", inputs.Node))
			return
		}
		// 删除a
		aPathLike := fmt.Sprintf("%s/%%/%%", pathTreeNode.Path)
		dt = db.Table(viper.GetString("table_path_tree")).Where(map[string]interface{}{"level": common.NodeLevelMap["a"]}).Where("path LIKE ?", aPathLike).Delete(common.PathTree{})
		if dt.Error != nil {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"del a under g failed:%s", dt.Error))
			return
		}

		// 	删除p
		pPathLike := fmt.Sprintf("%s/%%", pathTreeNode.Path)
		dt = db.Table(viper.GetString("table_path_tree")).Where(map[string]interface{}{"level": common.NodeLevelMap["p"]}).Where("path LIKE ?", pPathLike).Delete(common.PathTree{})
		if dt.Error != nil {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"del p under g failed:%s", dt.Error))
			return
		}
		// 	删除g
		dt = db.Table(viper.GetString("table_path_tree")).Delete(&pathTreeNode)
		if dt.Error != nil {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"del g failed:%s", dt.Error))
			return
		}
		common.JSONR(c, fmt.Sprintf(
			"del g and its p,a  successfully:%s", inputs.Node))
		return

	case common.NodeLevelMap["p"]:
		//删除p，先删除旗下的a:
		//查询p
		var pathTreeNode common.PathTree
		p := ss[1]
		dt = db.Table(viper.GetString("table_path_tree")).Where(map[string]interface{}{"level": inputs.Level, "node_name": p}).First(&pathTreeNode)
		if dt.Error != nil && !dt.RecordNotFound() {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"query p failed:%s", dt.Error))
			return
		}
		if dt.RowsAffected == 0 {
			common.JSONR(c, http.StatusBadRequest, fmt.Sprintf(
				"p not exists no need to del:%s", inputs.Node))
			return
		}
		// p存在先删除旗下的a
		aPathLike := fmt.Sprintf("%s/%%", pathTreeNode.Path)
		dt = db.Table(viper.GetString("table_path_tree")).Where(map[string]interface{}{"level": common.NodeLevelMap["a"]}).Where("path LIKE ?", aPathLike).Delete(common.PathTree{})
		if dt.Error != nil {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"del a under p failed:%s", dt.Error))
			return
		}
		// 删除p
		dt = db.Table(viper.GetString("table_path_tree")).Delete(&pathTreeNode)
		if dt.Error != nil {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"del  p failed:%s", dt.Error))
			return
		}
		common.JSONR(c, fmt.Sprintf(
			"del p and its a  successfully:%s", inputs.Node))
		return

	case common.NodeLevelMap["a"]:
		//删除a
		a := ss[2]
		dt = db.Table(viper.GetString("table_path_tree")).Where(map[string]interface{}{"level": inputs.Level, "node_name": a}).Delete(common.PathTree{})
		if dt.Error != nil {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"delete a  failed:%s", dt.Error))
			return
		}

	}

}

func GetGroupChildTree(c *gin.Context) {
	q := c.DefaultQuery("node", "")
	level := c.DefaultQuery("level", "")
	maxDep := c.DefaultQuery("max_dep", "0")
	if q == "" || level == "" {
		common.JSONR(c, http.StatusBadRequest, "node or level not provide")
		return
	}

	lev, err := strconv.ParseUint(level, 10, 64)
	if err != nil {
		common.JSONR(c, http.StatusBadRequest, "level  must be int ")
		return
	}

	maxD, err := strconv.ParseUint(maxDep, 10, 64)
	if err != nil {
		common.JSONR(c, http.StatusBadRequest, "maxDep must be int")
		return
	}

	inputs := common.NodeQueryReq{
		Level:  lev,
		Node:   q,
		MaxDep: maxD,
	}
	// 判断gpa是否合理
	if _, found := common.NodeLevelReverseMap[inputs.Level]; !found {
		common.JSONR(c, http.StatusBadRequest, fmt.Sprintf("level invalid:%d", inputs.Level))
		return
	}
	ss := strings.Split(inputs.Node, ".")
	ll := len(ss)

	if inputs.Level > 1 && int(inputs.Level) != ll+1 {
		common.JSONR(c, http.StatusBadRequest, fmt.Sprintf("node does not match  level:%d %s", inputs.Level, inputs.Node))
		return
	}

	var dt *gorm.DB
	var pTn common.PathTree
	var pTns []common.PathTree
	targets := make([]string, 0)
	db := pkg.GetDbCon()

	if inputs.MaxDep == 1 {
		// 只获取一级子节点
		q := ss[len(ss)-1]
		dt = db.Table(viper.GetString("table_path_tree")).Where(map[string]interface{}{"level": inputs.Level, "node_name": q}).Find(&pTn)
		if dt.Error != nil && !dt.RecordNotFound() {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"get child failed:%s", dt.Error))
			return
		}

		if dt.RowsAffected == 0 {
			common.JSONR(c, targets)
			return
		}
		cLike := fmt.Sprintf("%s/%%", pTn.Path)
		dt = db.Table(viper.GetString("table_path_tree")).Where(map[string]interface{}{"level": inputs.Level + 1}).Where("path LIKE ?", cLike).Find(&pTns)
		if dt.Error != nil && !dt.RecordNotFound() {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"get child failed:%s", dt.Error))
			return
		}
		for _, i := range pTns {
			targets = append(targets, i.NodeName)
		}

		sort.Strings(targets)

		common.JSONR(c, targets)
		return

	}

	switch inputs.Level {
	case common.NodeLevelMap["o"]:

		dt = db.Table(viper.GetString("table_path_tree")).Find(&pTns)
		if dt.Error != nil && !dt.RecordNotFound() {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"get all pga failed:%s", dt.Error))
			return
		}
		// 查询a
		existMapG := make(map[string]common.PathTree)
		existMapP := make(map[string]common.PathTree)
		existMapPId := make(map[uint64]common.PathTree)
		existMapA := make(map[string]common.PathTree)
		existMapAId := make(map[uint64]common.PathTree)

		for _, ptn := range pTns {
			switch ptn.Level {
			case common.NodeLevelMap["g"]:
				existMapG[ptn.Path] = ptn

			case common.NodeLevelMap["p"]:
				existMapP[ptn.Path] = ptn
				existMapPId[ptn.ID] = ptn
			case common.NodeLevelMap["a"]:
				existMapA[ptn.Path] = ptn
				existMapAId[ptn.ID] = ptn

			}

		}
		m := make(map[string]struct{})
		for gPath, g := range existMapG {
			for pId, p := range existMapPId {
				keyP := fmt.Sprintf("%s/%d", gPath, pId)
				if _, loaded := existMapP[keyP]; loaded {
					for aId, a := range existMapAId {
						keyA := fmt.Sprintf("%s/%d", p.Path, aId)
						if _, loaded := existMapA[keyA]; loaded {
							m[fmt.Sprintf("%s.%s.%s", g.NodeName, p.NodeName, a.NodeName)] = struct{}{}

						}
					}
					continue
				}
			}
		}
		for k, _ := range m {
			targets = append(targets, k)
		}
		sort.Strings(targets)
		common.JSONR(c, targets)
		return

	case common.NodeLevelMap["g"]:
		g := ss[0]
		dt = db.Table(viper.GetString("table_path_tree")).Where(map[string]interface{}{"level": inputs.Level, "node_name": g}).First(&pTn)
		if dt.Error != nil && !dt.RecordNotFound() {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"get g failed:%s", dt.Error))
			return
		}
		if dt.RowsAffected == 0 {
			common.JSONR(c, http.StatusBadRequest, fmt.Sprintf(
				"g not exists:%s ", inputs.Node))
			return
		}

		dt = db.Table(viper.GetString("table_path_tree")).Where("level IN (?)", []uint64{common.NodeLevelMap["p"], common.NodeLevelMap["a"]}).Find(&pTns)
		if dt.Error != nil {

			common.JSONR(c, http.StatusInternalServerError, dt.Error)
			return
		}

		existMapP := make(map[string]common.PathTree)
		existMapA := make(map[uint64]common.PathTree)

		for _, ptn := range pTns {
			switch ptn.Level {

			case common.NodeLevelMap["p"]:
				key := fmt.Sprintf("%s/%d", pTn.Path, ptn.ID)
				if key == ptn.Path {
					existMapP[ptn.Path] = ptn
				}

			case common.NodeLevelMap["a"]:
				if strings.Contains(ptn.Path, fmt.Sprintf("%s/", pTn.Path)) {
					existMapA[ptn.ID] = ptn
				}

			}

		}

		if len(existMapP) == 0 {
			common.JSONR(c, targets)
			return
		}
		for pPath, p := range existMapP {

			for aId, a := range existMapA {
				aPath := fmt.Sprintf("%s/%d", pPath, aId)
				if aPath == a.Path {
					targets = append(targets, fmt.Sprintf("%s.%s.%s", g, p.NodeName, a.NodeName))
				}

			}
		}
		sort.Strings(targets)
		common.JSONR(c, targets)
		return

	case common.NodeLevelMap["p"]:
		p := ss[1]
		dt = db.Debug().Table(viper.GetString("table_path_tree")).Where(map[string]interface{}{"level": inputs.Level, "node_name": p}).First(&pTn)
		if dt.Error != nil && !dt.RecordNotFound() {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"get p failed:%s", dt.Error))
			return
		}
		if dt.RowsAffected == 0 {
			common.JSONR(c, http.StatusBadRequest, fmt.Sprintf(
				"p not exists:%s ", inputs.Node))
			return
		}
		// 查询a
		aLike := fmt.Sprintf("%s/%%", pTn.Path)
		fmt.Println(aLike,pTn)
		dt = db.Debug().Table(viper.GetString("table_path_tree")).Where(map[string]interface{}{"level": common.NodeLevelMap["a"]}).Where("path LIKE ?", aLike).Find(&pTns)
		if dt.Error != nil && !dt.RecordNotFound() {
			common.JSONR(c, http.StatusInternalServerError, fmt.Sprintf(
				"get a under g.p failed:%s", dt.Error))
			return
		}
		for _, a := range pTns {
			targets = append(targets, fmt.Sprintf("%s.%s", inputs.Node, a.NodeName))
		}
		sort.Strings(targets)
		common.JSONR(c, targets)
		return

	}

}
