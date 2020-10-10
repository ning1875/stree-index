package main

import (
	"os"
	"os/signal"
	"context"
	"syscall"

	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/oklog/run"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"

	"stree-index/pkg"
	"stree-index/pkg/config"
	"stree-index/pkg/web"
	"stree-index/pkg/statistics"
	"stree-index/pkg/mem-index"
)

func main() {

	var (
		configFile = kingpin.Flag("config.file", "stree-index configuration file path.").Default("stree-index.yml").String()
	)

	// init logger
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("stree-index"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	// new ctxall
	ctxAll, cancelAll := context.WithCancel(context.Background())
	// init config
	sc, err := config.LoadFile(*configFile, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Load Config File Failed ....")
		return
	}
	// 为gin handlefunc 用到的变量设置一下
	config.SetDefaultVar(sc)

	// Init db && redis
	pkg.InitStore(sc)
	level.Info(logger).Log("msg", "Init Db and Redis Success....")

	pkg.NewMetrics()
	// index flush chan
	// 在索引刷新完成前不要启动web
	//indexUpdateChan := make(chan struct{}, 1)
	mem_index.InitIdx()
	err = mem_index.FlushAllIdx(logger)

	if err != nil {
		level.Error(logger).Log("msg", "Flush Index Failed And Exiting....")
		return
	}
	level.Info(logger).Log("msg", "Flush All Index Successfully ")
	var g run.Group
	{
		// Termination handler.
		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)
		cancel := make(chan struct{})
		g.Add(

			func() error {
				select {
				case <-term:
					level.Warn(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
					cancelAll()
					return nil
					//TODO clean work here
				case <-cancel:
					level.Warn(logger).Log("msg", "server finally exit...")
					return nil
				}
			},
			func(err error) {
				close(cancel)

			},
		)
	}

	{
		// metrics web handler.
		g.Add(func() error {
			//<-indexFlushDoneChan
			level.Info(logger).Log("msg", "start web service Listening on address", "address", sc.Http.ListenAddr)
			//gin.SetMode(gin.ReleaseMode)
			//routes := gin.Default()
			//routes := gin.New()
			errchan := make(chan error)

			go func() {
				errchan <- web.StartGin(sc.Http.ListenAddr)
			}()
			select {
			case err := <-errchan:
				level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
				return err
			case <-ctxAll.Done():
				level.Info(logger).Log("msg", "Web service Exit..")
				return nil

			}

		}, func(err error) {
			cancelAll()
		})
	}

	{
		// RegisterHeartBeatToCache   manager.
		g.Add(func() error {
			err := pkg.RegisterHeartBeatToCache(ctxAll, sc.RedisServer.Keys, logger)
			if err != nil {
				level.Error(logger).Log("msg", "RegisterHeartBeatToCache  error", "error", err)
			}

			return err
		}, func(err error) {
			cancelAll()
		})
	}

	{
		// IndexUpdate    manager.
		g.Add(func() error {
			err := mem_index.UpdateIndex(ctxAll, logger, sc.RedisServer.Keys)
			if err != nil {
				level.Error(logger).Log("msg", "IndexUpdateManager  error", "error", err)
			}

			return err
		}, func(err error) {
			cancelAll()
		})
	}

	{
		// Flush ALl IndexUpdate    manager.
		// cron flush all index in case of data Inconsistent
		g.Add(func() error {
			err := mem_index.FlushAllIndex(ctxAll, logger)
			if err != nil {
				level.Error(logger).Log("msg", "FluashALlIndexUpdateManager  error", "error", err)
			}

			return err
		}, func(err error) {
			cancelAll()
		})
	}

	{
		// TreeNodeStatistics    manager.
		g.Add(func() error {
			err := statistics.TreeNodeStatistics(ctxAll, logger)
			if err != nil {
				level.Error(logger).Log("msg", "TreeNodeStatistics  error", "error", err)
			}

			return err
		}, func(err error) {
			cancelAll()
		})
	}
	g.Run()
}
