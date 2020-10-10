package common

import (
	"fmt"
	"strings"
)

type QueryReq struct {
	ResourceType string          `json:"resource_type" binding:"required"`
	Labels       []*SingleTagReq `json:"labels" binding:"required"`
	UseIndex     bool            `json:"use_index" binding:"required"`
	TargetLabel  string          `json:"target_label" `
}

type QueryResponse struct {
	Code        int `json:"code"`
	CurrentPage int `json:"current_page"`
	PageSize    int `json:"page_size"`
	PageCount   int `json:"page_count"`
	TotalCount  int `json:"total_count"`
	Result   interface{} `json:"result"`
}

type SingleTagReq struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
	Type  int    `json:"type" binding:"required"`
}

func (q *QueryReq) PrintReq() string {
	keys := make([]string, 0)
	for _, i := range q.Labels {
		qType := ""
		switch i.Type {
		case 1:
			qType = "="
		case 2:
			qType = "!="
		case 3:
			qType = "~"
		case 4:
			qType = "!~"
		default:
			qType = "wrong"
		}

		keys = append(keys, fmt.Sprintf("%s%s%s", i.Key, qType, i.Value))
	}
	return strings.Join(keys, " ") + fmt.Sprintf(" target_label=%s", q.TargetLabel)
}
