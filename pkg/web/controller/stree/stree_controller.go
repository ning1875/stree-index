package stree

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	log "github.com/sirupsen/logrus"

	"stree-index/pkg"
	"stree-index/pkg/common"
	"stree-index/pkg/inverted-index"
	"stree-index/pkg/inverted-index/index"
	"stree-index/pkg/mem-index"
	"fmt"
	"strconv"
	"math"
)

func GetIndex(inputs common.QueryReq) (matchIds []uint64) {
	// get matcher
	matcher := common.FormatLabelMatchers(inputs.Labels)

	switch inputs.ResourceType {
	case "ecs":
		p, err := inverted_index.PostingsForMatchers(mem_index.EcsIdx, matcher...)
		if err != nil {
			log.Error("json_load_session_error", err)

		}
		matchIds, err = index.ExpandPostings(p)
	case "elb":
		p, err := inverted_index.PostingsForMatchers(mem_index.ElbIdx, matcher...)
		if err != nil {
			log.Error("json_load_session_error", err)

		}
		matchIds, err = index.ExpandPostings(p)
	case "rds":
		p, err := inverted_index.PostingsForMatchers(mem_index.RdsIdx, matcher...)
		if err != nil {
			log.Error("json_load_session_error", err)

		}
		matchIds, err = index.ExpandPostings(p)
	case "dcs":
		p, err := inverted_index.PostingsForMatchers(mem_index.DcsIdx, matcher...)
		if err != nil {
			log.Error("json_load_session_error", err)

		}
		matchIds, err = index.ExpandPostings(p)
	}
	return

}

func SelectEcsByIds(matchIds []uint64, limit, offset int) (ecsS []*common.Ecs) {
	db := pkg.GetDbCon()
	var dt *gorm.DB

	if limit == 0 && offset == 0 {
		dt = db.Debug().Table(viper.GetString("table_ecs")).Where(matchIds).Find(&ecsS)
	} else {
		dt = db.Debug().Table(viper.GetString("table_ecs")).Where(matchIds).Limit(limit).Offset(offset).Find(&ecsS)
	}

	if dt.Error != nil {

		log.Errorf("find ecs failed", dt.Error)
		return nil
	}
	return
}

func SelectLbByIds(matchIds []uint64, limit, offset int) (elbs []*common.Elb) {
	db := pkg.GetDbCon()
	var dt *gorm.DB
	if limit == 0 && offset == 0 {
		dt = db.Debug().Table(viper.GetString("table_elb")).Where(matchIds).Find(&elbs)
	} else {
		dt = db.Debug().Table(viper.GetString("table_elb")).Where(matchIds).Limit(limit).Offset(offset).Find(&elbs)
	}

	if dt.Error != nil {

		log.Errorf("find elb failed", dt.Error)
		return nil
	}
	return
}

func SelectRdsByIds(matchIds []uint64, limit, offset int) (rdsS []*common.Rds) {
	db := pkg.GetDbCon()
	var dt *gorm.DB
	if limit == 0 && offset == 0 {
		dt = db.Debug().Table(viper.GetString("table_rds")).Where(matchIds).Find(&rdsS)
	} else {
		dt = db.Debug().Table(viper.GetString("table_rds")).Where(matchIds).Limit(limit).Offset(offset).Find(&rdsS)

	}

	if dt.Error != nil {

		log.Errorf("find rds failed", dt.Error)
		return nil
	}
	return
}

func SelectDcsByIds(matchIds []uint64, limit, offset int) (items []*common.Dcs) {
	db := pkg.GetDbCon()
	var dt *gorm.DB
	if limit == 0 && offset == 0 {
		dt = db.Debug().Table(viper.GetString("table_dcs")).Where(matchIds).Find(&items)
	}else {
		dt = db.Debug().Table(viper.GetString("table_dcs")).Where(matchIds).Limit(limit).Offset(offset).Find(&items)
	}


	if dt.Error != nil {

		log.Errorf("find dcs failed", dt.Error)
		return nil
	}
	return
}

func GetStats(c *gin.Context) {
	qType := c.DefaultQuery("type", "ecs")

	if _, ok := mem_index.SupportIndex[qType]; !ok {
		common.JSONR(c, http.StatusBadRequest, "unsport index type:"+qType)
		return
	}

	switch qType {
	case "ecs":
		res := mem_index.EcsIdx.PostingsCardinalityStats()
		common.JSONR(c, res)
		return
	}

}

func GetLabelGroup(c *gin.Context) {
	qType := c.DefaultQuery("type", "ecs")
	label := c.DefaultQuery("label", "region")

	if _, ok := mem_index.SupportIndex[qType]; !ok {
		common.JSONR(c, http.StatusBadRequest, "unsport index type:"+qType)
		return
	}

	switch qType {
	case "ecs":
		res := mem_index.EcsIdx.GetGroupByLabel(label)
		common.JSONR(c, res)
		return
	}

}

func GetLabelDistribution(c *gin.Context) {
	var inputs common.QueryReq
	if err := c.Bind(&inputs); err != nil {
		common.JSONR(c, http.StatusBadRequest, err)
		return
	}
	// check for regex *
	for _, x := range inputs.Labels {
		invalid := false
		if x.Type == 3 {
			switch x.Value {
			case "*":
				invalid = true
			}
		}
		if invalid {
			common.JSONR(c, http.StatusBadRequest, "* regex value not allow")
			return
		}

	}

	if _, ok := mem_index.SupportIndex[inputs.ResourceType]; !ok {
		common.JSONR(c, http.StatusBadRequest, "unsport index type:"+inputs.ResourceType)
		return
	}

	matchIds := GetIndex(inputs)
	log.Infof("[GetLabelDistributionGetIdRes][inputs.Labels:%+v][matchIds:%+v]", inputs.PrintReq(), matchIds)
	if len(matchIds) == 0 {

		res := inverted_index.NewLabelGroup()
		res.Message = "found zero record by given matchers"
		common.JSONR(c, res)
		return
	}

	switch inputs.ResourceType {
	case "ecs":
		res := mem_index.EcsIdx.GetGroupDistributionByLabel(inputs.TargetLabel, matchIds)
		common.JSONR(c, res)
		return
	case "elb":
		res := mem_index.ElbIdx.GetGroupDistributionByLabel(inputs.TargetLabel, matchIds)
		common.JSONR(c, res)
		return

	case "rds":
		res := mem_index.RdsIdx.GetGroupDistributionByLabel(inputs.TargetLabel, matchIds)
		common.JSONR(c, res)
		return
	case "dcs":
		res := mem_index.DcsIdx.GetGroupDistributionByLabel(inputs.TargetLabel, matchIds)
		common.JSONR(c, res)
		return
	}

}

func GetRdsTest(c *gin.Context) {
	db := pkg.GetDbCon()
	var dt *gorm.DB
	var rdsS []common.Rds
	dt = db.Table(viper.GetString("table_rds")).Where(map[string]interface{}{"uid": "62787405f3224aee8aa7985a0ea8384ein01"}).Find(&rdsS)

	if dt.Error != nil {

		log.Errorf("find rds failed", dt.Error)
		return
	}
	c.JSON(200, rdsS)
	return
}

func GetResourceByKeyV(c *gin.Context) {

	var inputs common.QueryReq
	if err := c.Bind(&inputs); err != nil {
		common.JSONR(c, http.StatusBadRequest, err)
		return
	}
	if !inputs.UseIndex {
		//TODO 不使用索引用raw sql
		common.JSONR(c, http.StatusBadRequest, "raw sql not support")
		return
	}
	// check for regex *
	for _, x := range inputs.Labels {
		invalid := false
		if x.Type == 3 {
			switch x.Value {
			case "*":
				invalid = true
			}
		}
		if invalid {
			common.JSONR(c, http.StatusBadRequest, "* regex value not allow")
			return
		}

	}

	if _, ok := mem_index.SupportIndex[inputs.ResourceType]; !ok {
		common.JSONR(c, http.StatusBadRequest, "unsport index type:"+inputs.ResourceType)
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		common.JSONR(c, http.StatusBadRequest, "invalid page_size:")
		return
	}

	currentPage, err := strconv.Atoi(c.DefaultQuery("current_page", "1"))
	if err != nil {
		common.JSONR(c, http.StatusBadRequest, "invalid current_page:")
		return
	}

	getAll, err := strconv.Atoi(c.DefaultQuery("get_all", "0"))
	if err != nil {
		common.JSONR(c, http.StatusBadRequest, "invalid getAll:")
		return
	}

	var offset int = 0
	var limit int = 0
	limit = pageSize
	if currentPage > 1 {
		offset = (currentPage - 1) * limit
	}

	matchIds := GetIndex(inputs)

	totalCount := len(matchIds)
	PageCount := int(math.Ceil(float64(totalCount) / float64(limit)))
	resp := common.QueryResponse{}
	resp.PageSize = pageSize
	resp.PageCount = PageCount
	resp.CurrentPage = currentPage
	resp.TotalCount = totalCount

	if getAll == 1 {
		limit = 0
		offset = 0
	}
	log.Infof("[GetResourceByKeyVGetIdRes][inputs.Labels:%+v][matchIds:%+v]", inputs.PrintReq(), matchIds)
	switch inputs.ResourceType {
	case "ecs":
		rs := SelectEcsByIds(matchIds, limit, offset)

		resp.Result = rs

		common.JSONR(c, resp)
		return
	case "elb":
		rs := SelectLbByIds(matchIds, limit, offset)
		resp.Result = rs
		common.JSONR(c, resp)
		return
	case "rds":
		rs := SelectRdsByIds(matchIds, limit, offset)
		resp.Result = rs
		common.JSONR(c, resp)
		return
	case "dcs":
		rs := SelectDcsByIds(matchIds, limit, offset)
		resp.Result = rs
		common.JSONR(c, resp)
		return
	}

	return

}

func IndexUpdate(c *gin.Context) {
	resourceType := c.DefaultQuery("resource_type", "")

	apiToken := c.Request.Header.Get(viper.GetString("api_token_key"))

	allowToken := viper.GetString("api_token")
	if apiToken == "" {

		common.JSONR(c, http.StatusUnauthorized, "token key is not set ")
		return
	}

	if allowToken != "" && apiToken != allowToken {
		fmt.Println(apiToken, allowToken)
		common.JSONR(c, http.StatusUnauthorized, "wrong token key")
		return
	}
	if _, ok := mem_index.SupportIndex[resourceType]; !ok {
		common.JSONR(c, http.StatusBadRequest, "unsport index type:"+resourceType)
		return
	}

	log.Printf("[IndexUpdate]called by %s resourceType:%s", c.Request.RemoteAddr, resourceType)
	pkg.IndexUpdateChan <- resourceType
	common.JSONR(c, "start update index for  resourceType: "+resourceType)
	return
}
