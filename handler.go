package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/jlaffaye/ftp"
	"github.com/julienschmidt/httprouter"
	"github.com/mholt/archiver/v3"
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

func getRegion(regionId int) (configRegion, error) {
	var region configRegion
	index := -1

	for i, item := range appConfig.Regions {
		if item.RegionId == regionId {
			index = i
		}
	}

	if index == -1 {
		return region, errors.New("未找到 region")
	}

	region = appConfig.Regions[index]
	return region, nil
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

	region, err := getRegion(body.RegionId)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(res, "invalid regionId: %d\n", body.RegionId)
		return body, region
	}

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
	runBat(res, req, "📢开服", func(region configRegion) string { return region.Start })
}

func stopHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	runBat(res, req, "⚙️关服", func(region configRegion) string { return region.Stop })
}

func updateServerHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	_, region := getReqBodyAndRegion(res, req)
	if region == (configRegion{}) {
		res.WriteHeader(http.StatusBadRequest)
		msg := "❓未找到对应的 Region"
		fmt.Fprintln(res, msg)
		log.Println(msg)
		return
	}

	dirPath := path.Join(appConfig.Ftp.Path, "Game")
	name, err := downloadFromFtp(dirPath, region)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		log.Println(err)
		return
	}

	log.Printf("⏳正在解压📦'%s'……\n", name)
	err = archiver.Unarchive(name, region.WorkDir)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		log.Println(err)
		return
	}

	fmt.Fprintln(res, "✔️")
}

func updateConfigHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	_, region := getReqBodyAndRegion(res, req)
	if region == (configRegion{}) {
		res.WriteHeader(http.StatusBadRequest)
		msg := "❓未找到对应的 Region"
		fmt.Fprintln(res, msg)
		log.Println(msg)
		return
	}

	dirPath := path.Join(appConfig.Ftp.Path, "Config")
	name, err := downloadFromFtp(dirPath, region)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		log.Println(err)
		return
	}

	log.Printf("⏳正在解压📦'%s'……\n", name)
	err = archiver.Unarchive(name, region.WorkDir)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		log.Println(err)
		return
	}

	fmt.Fprintln(res, "✔️")
}

func downloadFromFtp(dirPath string, region configRegion) (string, error) {
	host := appConfig.Ftp.Host
	host = strings.TrimPrefix(host, "ftp://")
	host = strings.TrimPrefix(host, "ftps://")

	addr := fmt.Sprintf("%s:%d", host, appConfig.Ftp.Port)
	conn, err := ftp.Dial(addr)
	if err != nil {
		return "", err
	}
	err = conn.Login(appConfig.Ftp.User, appConfig.Ftp.Password)
	if err != nil {
		return "", err
	}

	entries, err := conn.List(dirPath)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", errors.New("❌更新失败，未找到更新包。")
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name > entries[j].Name
	})

	name := path.Join(dirPath, entries[0].Name)
	log.Printf("⏳正在下载'%s'……\n", name)
	res, err := conn.Retr(name)
	if err != nil {
		return "", err
	}
	defer res.Close()

	buf, err := ioutil.ReadAll(res)
	if err != nil {
		return "", err
	}
	newName := path.Join(region.WorkDir, entries[0].Name)
	ioutil.WriteFile(newName, buf, 0644)
	log.Println("✔️下载完成")

	return newName, nil
}

func dmpHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	strRegionId := req.URL.Query().Get("regionId")
	regionId, err := strconv.Atoi(strRegionId)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		msg := "❌regionId 不合符"
		fmt.Fprintln(res, msg)
		log.Println(msg)
		return
	}

	region, err := getRegion(regionId)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		msg := "❓未找到对应的 Region"
		fmt.Fprintln(res, msg)
		log.Println(msg)
		return
	}

	files, err := ioutil.ReadDir(region.WorkDir)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		log.Println(err)
		return
	}
	count := 0
	for _, file := range files {
		ext := path.Ext(strings.ToLower(file.Name()))
		if !file.IsDir() && ext == ".dmp" {
			count++
		}
	}

	fmt.Fprintln(res, count)
}

func ipFilter(handler httprouter.Handle) httprouter.Handle {
	return func(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
		ip := req.Header.Get("X-Real-Ip")
		if ip == "" {
			ip = req.Header.Get("X-Forwarded-For")
		}
		if ip == "" {
			ip = req.RemoteAddr
		}
		ip = strings.Split(ip, ":")[0]

		contains := false
		for _, item := range appConfig.Whitelist {
			if item == ip {
				contains = true
				break
			}
		}

		if contains {
			handler(res, req, p)
		} else {
			res.WriteHeader(http.StatusForbidden)
			msg := fmt.Sprintf("非白名单 ip '%s' 试图访问服务，已拒绝。", ip)
			fmt.Fprintln(res, msg)
			log.Println(msg)
		}
	}
}
