package node_path

import (
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.Engine) {
	/*
	1.给lb的探活接口
	2.根据tag kv 查询资源的接口
	*/

	pApi := r.Group("/stree-index/node-path")
	//pApi.POST("/*any", AddNode)
	//pApi.GET("/*any", GetGroupChildTree)
	pApi.POST("/*any", AddNodeByGetAll)
	pApi.DELETE("/*any", DeleteNode)

}
