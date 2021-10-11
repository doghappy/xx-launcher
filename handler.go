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

func StartHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(req.Body)
	var body reqBody

	if err := decoder.Decode(&body); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, err.Error())
		return
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
		return
	}

	log.Printf("[开服] RegionId: %d\n", body.RegionId)

	region := appConfig.Regions[index]
	name := path.Join(region.WorkDir, region.Start)

	_, err := startProcess(name)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		log.Println(err)
		return
	}

	fmt.Fprintln(res, "ok")
}
