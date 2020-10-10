package common

import (
	"time"
	"gorm.io/datatypes"
)

type Common struct {
	// common 13个
	ID            uint64 `json:"id"                 gorm:"column:id"`
	Uid           string `json:"uid"                gorm:"column:uid"`  //uuid
	Hash          string `json:"hash"               gorm:"column:hash"` //所有key v的hash，判读其是否改变
	Name          string `json:"name"               gorm:"column:name"`
	CloudProvider string `json:"cloud_provider"     gorm:"column:cloud_provider"` // 公有云厂商
	ChargingMode  string `json:"charging_mode"      gorm:"column:charging_mode"`  // 付费模式
	Region        string `json:"region"             gorm:"column:region"`         // region
	AccountId     uint64 `json:"account_id"         gorm:"column:account_id"`
	VpcId         string `json:"vpc_id"             gorm:"column:vpc_id"` // vpc id
	SubnetId      string `json:"subnet_id"          gorm:"column:subnet_id"`
	//SecurityGroups string    `json:"security_groups"    gorm:"column:security_groups"`  // 绑定的安全组
	SecurityGroups datatypes.JSON `json:"security_groups"    gorm:"column:security_groups"` // 绑定的安全组
	PrivateIp      datatypes.JSON `json:"private_ip"         gorm:"column:private_ip"`      // 内网ips
	Status         string         `json:"status"              gorm:"column:status"`         // 状态
	//Tags           string    `json:"tags"               gorm:"column:tags"`             // 标签json
	Tags      datatypes.JSON `json:"tags"               gorm:"column:tags"`             // 标签json
	CreatedAt time.Time      `json:"created_at"               gorm:"column:created_at"` // 标签json
}

type Ecs struct {
	Common

	// 独有
	OwnerId      string `json:"owner_id"           gorm:"column:owner_id"`
	InstanceType string `json:"instance_type"      gorm:"column:instance_type"` // 规格

	PublicIp         datatypes.JSON `json:"public_ip"          gorm:"column:public_ip"`         // 公网ips
	AvailabilityZone string         `json:"availability_zone"  gorm:"column:availability_zone"` // 可用区
}

type Elb struct {
	Common
	ElbType     string         `json:"elb_type"      gorm:"column:elb_type"`         // 规格
	IpAddress   string         `json:"ip_address"      gorm:"column:ip_address"`     // 规格
	DnsName     string         `json:"dns_name"      gorm:"column:dns_name"`         // 规格
	Backends    datatypes.JSON `json:"backends"      gorm:"column:backends"`         // 规格
	Port        datatypes.JSON `json:"port"      gorm:"column:port"`                 // 规格
	TargetGroup datatypes.JSON `json:"target_group"      gorm:"column:target_group"` // 规格
}

type Rds struct {
	// common 13个
	Common

	// 独有
	Engine           string `json:"engine"      gorm:"column:engine"`                       // mysql or postgresql
	EngineVersion    string `json:"engine_version"      gorm:"column:engine_version"`       // 版本号
	InstanceType     string `json:"instance_type"      gorm:"column:instance_type"`         //规格
	ArchitectureType string `json:"architecture_type"      gorm:"column:architecture_type"` // ha or single
	//Flavor        string `json:"flavor"      gorm:"column:flavor"`                 //
	ClusterId   uint64         `json:"cluster_id"      gorm:"column:cluster_id"`       // 规格
	ClusterName string         `json:"cluster_name"      gorm:"column:cluster_name"`   // 规格
	MasterId    string         `json:"master_id"      gorm:"column:master_id"`         // 规格
	PublicIp    datatypes.JSON `json:"public_ip"          gorm:"column:public_ip"`     // 公网ips
	ResourceId  string         `json:"resource_id"          gorm:"column:resource_id"` // resource_id
	IsWriter    bool           `json:"is_writer"          gorm:"column:is_writer"`     // 是否读写
	Port        uint64         `json:"port"                 gorm:"column:port"`        // 端口
}

type Dcs struct {
	Common
	ClusterId     string `json:"cluster_id"      gorm:"column:cluster_id"`         // 规格
	InstanceType  string `json:"instance_type"      gorm:"column:instance_type"`   // ha or single
	PublicIp      string `json:"public_ip"          gorm:"column:public_ip"`       // 公网ips
	Port          uint64 `json:"port"                 gorm:"column:port"`          // 端口
	Engine        string `json:"engine"      gorm:"column:engine"`                 // mysql or postgresql
	EngineVersion string `json:"engine_version"      gorm:"column:engine_version"` // 版本号
}

type PathTree struct {
	ID       uint64 `json:"id"                 gorm:"column:id"`
	Level    uint64 `json:"level"           gorm:"column:level"`
	Path     string `json:"path"           gorm:"column:path"`
	NodeName string `json:"node_name"           gorm:"column:node_name"`
}

type InstanceType struct {
	ID           uint64 `json:"id"                 gorm:"column:id"`
	ResourceType string `json:"resource_type" gorm:"resource_type"`

	InstanceType  string `json:"instance_type" gorm:"instance_type"`
	CloudProvider string `json:"cloud_provider" gorm:"cloud_provider"`
	Cpu           uint64 `json:"cpu" gorm:"cpu"`
	Mem           uint64 `json:"mem" gorm:"mem"`
}

type IndexIncrementUpdate struct {
	ToAdd     []uint64 `json:"to_add"`
	ToDel     []uint64 `json:"to_del"`
	ToMod     []uint64 `json:"to_mod"`
	ToDelHash []string `json:"to_del_hash"`
}
