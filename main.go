package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
	"github.com/prometheus/common/promlog"
	promlogflag "github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"stree-index/pkg"
	"stree-index/pkg/config"
	"stree-index/pkg/mem-index"
	"stree-index/pkg/statistics"
	"stree-index/pkg/web"
)

func main() {

	var (
		app = kingpin.New(filepath.Base(os.Args[0]), "The stree-index")
		//configFile = kingpin.Flag("config.file", "docker-mon configuration file path.").Default("docker-mon.yml").String()
		configFile = app.Flag("config.file", "docker-mon configuration file path.").Default("stree-index.yml").String()
	)
	promlogConfig := promlog.Config{}

	app.Version(version.Print("stree-index"))
	app.HelpFlag.Short('h')
	promlogflag.AddFlags(app, &promlogConfig)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	var logger log.Logger
	logger = func(config *promlog.Config) log.Logger {
		var (
			l  log.Logger
			le level.Option
		)
		if config.Format.String() == "logfmt" {
			l = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		} else {
			l = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		}

		switch config.Level.String() {
		case "debug":
			le = level.AllowDebug()
		case "info":
			le = level.AllowInfo()
		case "warn":
			le = level.AllowWarn()
		case "error":
			le = level.AllowError()
		}
		l = level.NewFilter(l, le)
		l = log.With(l, "ts", log.TimestampFormat(
			func() time.Time { return time.Now().Local() },
			"2006-01-02T15:04:05.000Z07:00",
		), "caller", log.DefaultCaller)
		return l
	}(&promlogConfig)

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
