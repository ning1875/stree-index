package common

type NodeAddReq struct {
	Node  string `json:"node"`
	Level uint64 `json:"level"`
}

type NodeQueryReq struct {
	Node   string `json:"node"`
	Level  uint64 `json:"level"`
	MaxDep uint64 `json:"max_dep"`
}
