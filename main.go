package main

import (
	"github.com/JunxiHe459/gateway/global"
	"github.com/JunxiHe459/gateway/router"
	"github.com/e421083458/golang_common/lib"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	initConf()
	initDB()
}

func main() {
	defer lib.Destroy()
	router.HttpServerRun()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	router.HttpServerStop()
}

func initConf() {
	_ = lib.InitModule("./conf/dev/", []string{"base", "mysql", "redis"})
}

func initDB() {
	var err error
	global.DB, err = lib.GetGormPool("default")
	global.DB = global.DB.Debug()
	if err != nil {
		print("Get Global Variable DB failed")
	}
}
