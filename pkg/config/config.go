package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"
)

type Config struct {
	MysqlServer *MysqlServerConfig `yaml:"mysql_server"`
	RedisServer *RedisServerConfig `yaml:"redis_server"`
	Http        *Http              `yaml:"http"`
	IndexUpdate *IndexUpdate       `yaml:"index_update"`
	TreeInfo    *TreeInfo          `yaml:"tree_info"`
}

type TreeInfo struct {
	NameG string `yaml:"name_g"`
	NameP string `yaml:"name_p"`
	NameA string `yaml:"name_a"`
}

type MysqlServerConfig struct {
	Host         string           `yaml:"host"`
	Username     string           `yaml:"username"`
	Password     string           `yaml:"password"`
	Dbname       string           `yaml:"dbname"`
	LogPrint     bool             `yaml:"log_print"`
	MaxIdleConns int              `yaml:"maxidleconns"`
	MaxOpenConns int              `yaml:"maxopenconns"`
	TableInfo    *TableInfoConfig `yaml:"table_info"`
}

type TableInfoConfig struct {
	TableEcs               string `yaml:"table_ecs"`
	TableElb               string `yaml:"table_elb"`
	TableRdb               string `yaml:"table_rds"`
	TableDcs               string `yaml:"table_dcs"`
	TableDns               string `yaml:"table_dns"`
	TablePathTree          string `yaml:"table_path_tree"`
	TableCloudInstanceType string `yaml:"table_cloud_instance_type"`
}

type RedisServerConfig struct {
	Host     string     `yaml:"host"`
	Password string     `yaml:"password,omitempty"`
	Dbname   int        `yaml:"dbname,omitempty"`
	Keys     *RedisKeys `yaml:"keys"`
}

type RedisKeys struct {
	IndexNodeMapName        string `yaml:"index_node_map_name"`
	IndexNodeKeypPrefix     string `yaml:"index_node_key_prefix"`
	IndexIncrementUpdateKey string `yaml:"index_increment_update_key"`
	AllGpaCacheKey          string `yaml:"all_gpa_cache_key"`
	ToDelIds                string `yaml:"to_del_ids"`
	ToModIds                string `yaml:"to_mod_ids"`
}

type Http struct {
	ListenAddr string `yaml:"listen_addr"`
}

type IndexUpdate struct {
	HeaderKey                    string `yaml:"header_key"`
	Token                        string `yaml:"token"`
	AllIndexUpdateIntervalMinute int    `yaml:"all_index_update_interval_minute"`
	TreeNodeStatisticsInterval   int    `yaml:"tree_node_statistics_interval_minute"`
}

//type IndexConfig struct {
//	SupportType []string `yaml:support_type,omitempty`
//}

func Load(s string) (*Config, error) {
	cfg := &Config{}

	err := yaml.UnmarshalStrict([]byte(s), cfg)

	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func LoadFile(filename string, logger log.Logger) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg, err := Load(string(content))
	if err != nil {
		level.Error(logger).Log("msg", "parsing YAML file errr...", "error", err)
		return nil, err
	}
	return cfg, nil
}

func SetDefaultVar(sc *Config) {

	if sc.IndexUpdate != nil {
		viper.SetDefault("api_token_key", sc.IndexUpdate.HeaderKey)
		viper.SetDefault("api_token", sc.IndexUpdate.Token)
		viper.SetDefault("all_index_update_interval_minute", sc.IndexUpdate.AllIndexUpdateIntervalMinute)
		viper.SetDefault("tree_node_statistics_interval_minute", sc.IndexUpdate.TreeNodeStatisticsInterval)
	} else {
		viper.SetDefault("api_token_key", "Authorization")
		viper.SetDefault("api_token", "Admin")
		viper.SetDefault("all_index_update_interval_minute", 60)
		viper.SetDefault("tree_node_statistics_interval_minute", 10)
	}

	if sc.MysqlServer.TableInfo != nil {
		viper.SetDefault("table_ecs", sc.MysqlServer.TableInfo.TableEcs)
		viper.SetDefault("table_elb", sc.MysqlServer.TableInfo.TableElb)
		viper.SetDefault("table_rds", sc.MysqlServer.TableInfo.TableRdb)
		viper.SetDefault("table_dcs", sc.MysqlServer.TableInfo.TableDcs)
		viper.SetDefault("table_path_tree", sc.MysqlServer.TableInfo.TablePathTree)
		viper.SetDefault("table_cloud_instance_type", sc.MysqlServer.TableInfo.TableCloudInstanceType)
	} else {
		viper.SetDefault("table_ecs", "ecs")
		viper.SetDefault("table_elb", "elb")
		viper.SetDefault("table_rds", "rds")
		viper.SetDefault("table_dcs", "dcs")
		viper.SetDefault("table_dns", "dns")
		viper.SetDefault("table_path_tree", "path_tree")
		viper.SetDefault("table_cloud_instance_type", "cloud_instance_type")
	}

	if sc.TreeInfo != nil {
		viper.SetDefault("tree_g_name", sc.TreeInfo.NameG)
		viper.SetDefault("tree_p_name", sc.TreeInfo.NameP)
		viper.SetDefault("tree_a_name", sc.TreeInfo.NameA)
	} else {
		viper.SetDefault("tree_g_name", "group")
		viper.SetDefault("tree_p_name", "product")
		viper.SetDefault("tree_a_name", "app")
	}

	if sc.RedisServer != nil {
		viper.SetDefault("all_gpa_cache_key", sc.RedisServer.Keys.AllGpaCacheKey)
	} else {
		viper.SetDefault("all_gpa_cache_key", "all_gpa_cache_key")
	}

	if sc.Http != nil {
		viper.SetDefault("http_listen_addr", sc.Http.ListenAddr)
	} else {
		viper.SetDefault("http_listen_addr", "9292")
	}
}
