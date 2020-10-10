package web

import (
	"time"
	"net/http"
	"github.com/gin-gonic/gin"

	"stree-index/pkg/web/controller/stree"
	"stree-index/pkg/web/controller/node-path"
)

func StartGin(port string) error {
	r := gin.New()

	p := NewPrometheus("gin")
	//gin.SetMode(gin.ReleaseMode)
	r.Use(gin.Logger())
	p.Use(r)
	stree.Routes(r)
	node_path.Routes(r)

	s := &http.Server{
		Addr:           port,
		Handler:        r,
		ReadTimeout:    time.Duration(5) * time.Second,
		WriteTimeout:   time.Duration(5) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServe()
	return err
}
