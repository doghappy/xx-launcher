package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/yaml.v2"
)

var appConfig = config{}

func main() {
	file, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Panicln("⚠️日志无法写入到文件！")
	}
	mw := io.MultiWriter(os.Stdout, file)
	defer file.Close()
	log.SetOutput(mw)
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	readConfig()
	router := httprouter.New()
	router.POST("/start", ipFilter(startHandler))
	router.POST("/stop", ipFilter(stopHandler))
	router.POST("/server", ipFilter(updateServerHandler))
	router.POST("/config", ipFilter(updateConfigHandler))
	router.GET("/dmp", ipFilter(dmpHandler))
	log.Printf("[启动器] 服务地址 %s\n", appConfig.LauncherUrl)
	log.Fatalln(http.ListenAndServe(appConfig.LauncherUrl, router))
}

func readConfig() {
	bytes, err := os.ReadFile("config.yml")
	if err != nil {
		log.Fatalln(err)
	}
	yaml.Unmarshal(bytes, &appConfig)
}
