package pkg

import (
	"fmt"
	"time"
	"log"

	"github.com/jinzhu/gorm"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"

	"stree-index/pkg/config"
	"stree-index/pkg/common"
)

const (
	RedisIndexNodeExpire  = time.Second * 65
	AliveRegisterInterval = time.Minute * 1
)

var (
	Db              *gorm.DB
	dbp             *DBPool
	cache           *redis.Client
	LocalIp         = common.GetLocalIp()
	IndexUpdateChan = make(chan string, 10)
)

type DBPool struct {
	Tree *gorm.DB
}

func GetDbCon() *gorm.DB {
	return Db
}
func GetRedis() *redis.Client {
	return cache
}

func InitStore(sc *config.Config) {
	InitDb(sc.MysqlServer)

	InitRedis(sc.RedisServer)
}

func InitDb(dbConfig *config.MysqlServerConfig) {
	uri := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Dbname)
	db, err := gorm.Open("mysql", uri)
	if err != nil {
		log.Fatalln("open mysql error and exit...", err)
	}

	if err := db.DB().Ping(); err != nil {
		log.Fatalln("ping mysql error and exit...", err)
	}

	db.LogMode(dbConfig.LogPrint)
	db.DB().SetMaxIdleConns(dbConfig.MaxIdleConns)
	db.DB().SetMaxOpenConns(dbConfig.MaxOpenConns)
	db.DB().SetConnMaxLifetime(time.Hour)
	Db = db
}

func InitRedis(cnf *config.RedisServerConfig) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cnf.Host,
		//Password: cnf.Password,
		DB: cnf.Dbname,
	})
	//ctx, _ := context.WithCancel(context.Background())
	_, err := rdb.Ping().Result()
	//_, err := rdb.Ping().Result()
	if err != nil {

		log.Fatalln("open  redis fail:", err)
	}
	cache = rdb
}
