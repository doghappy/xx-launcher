package main

import (
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/yaml.v2"
)

func init() {
	// log.SetPrefix("launcher: ")
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
}

var appConfig = config{}

func main() {
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
