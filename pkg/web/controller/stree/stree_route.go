package stree

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"stree-index/pkg"

	"stree-index/pkg/web/controller/node-path"
)

func Routes(r *gin.Engine) {
	/*
	1.给lb的探活接口
	2.根据tag kv 查询资源的接口
	*/

	qapi := r.Group("/stree-index/query")
	qapi.GET("/node-path", node_path.GetGroupChildTree)

	qapi.POST("/resource/*any", GetResourceByKeyV)

	qapi.POST("/resource-distribution/*any", GetLabelDistribution)

	qapi.GET("/stats/*any", GetStats)

	qapi.GET("/resource/group/*any", GetLabelGroup)
	// TODO 索引增量更新接口开关:来自sync-worker
	//upApi := r.Group("/index/update")
	//upApi.GET("/*any", IndexUpdate)

	//hapi := r.Group("/healthy")
	r.GET("/stree-index/healthy", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, I'm pgw stree-index+  from :"+pkg.LocalIp)
	})
}
