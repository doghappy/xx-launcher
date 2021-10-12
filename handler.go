package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/julienschmidt/httprouter"
)

type reqBody struct {
	RegionId int `json:"regionId"`
	Type     int `json:"type"`
}

var startProcess = func(name string) (*os.Process, error) {
	procAttr := &os.ProcAttr{
		Env: os.Environ(),
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	}
	return os.StartProcess(name, []string{}, procAttr)
}

func getReqBodyAndRegion(res http.ResponseWriter, req *http.Request) (reqBody, configRegion) {
	decoder := json.NewDecoder(req.Body)
	body := reqBody{}
	region := configRegion{}

	if err := decoder.Decode(&body); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, err.Error())
		return body, region
	}

	index := -1

	for i, item := range appConfig.Regions {
		if item.RegionId == body.RegionId {
			index = i
		}
	}

	if index == -1 {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(res, "invalid regionId: %d\n", body.RegionId)
		return body, region
	}

	region = appConfig.Regions[index]

	return body, region
}

func runBat(res http.ResponseWriter, req *http.Request, prefix string, nameProvider func(configRegion) string) {
	body, region := getReqBodyAndRegion(res, req)
	if body != (reqBody{}) && region != (configRegion{}) {
		log.Printf("[%s] RegionId: %d\n", prefix, body.RegionId)

		name := path.Join(region.WorkDir, nameProvider(region))

		_, err := startProcess(name)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(res, err.Error())
			log.Println(err)
			return
		}

		fmt.Fprintln(res, "ok")
	}
}

func startHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	runBat(res, req, "开服", func(region configRegion) string { return region.Start })
}

func stopHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	runBat(res, req, "关服", func(region configRegion) string { return region.Stop })
}
