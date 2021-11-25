package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/julienschmidt/httprouter"
	"github.com/mholt/archiver/v3"
)

type reqBody struct {
	RegionId int `json:"regionId"`
}

type ftpResult struct {
	file string
	err  error
}

var lock = sync.Mutex{}

var startProcess = func(dir string, name string) (*os.Process, error) {
	procAttr := &os.ProcAttr{
		Env: os.Environ(),
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	}
	procAttr.Dir = dir
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

		_, err := startProcess(region.WorkDir, nameProvider(region))
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(res, err.Error())
			log.Println(err)
			return
		}

		log.Printf("%s成功\n", prefix)

		fmt.Fprintln(res, "ok")
	}
}

func startHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	lock.Lock()
	defer lock.Unlock()
	runBat(res, req, "📢开服", func(region configRegion) string { return region.Start })
}

func stopHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	lock.Lock()
	defer lock.Unlock()
	runBat(res, req, "⚙️关服", func(region configRegion) string { return region.Stop })
}

func updateServerHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	lock.Lock()
	log.Println("🔒加锁")
	defer func() {
		log.Println("🔓解锁")
		lock.Unlock()
	}()

	_, region := getReqBodyAndRegion(res, req)
	if region == (configRegion{}) {
		res.WriteHeader(http.StatusBadRequest)
		msg := "❓未找到对应的 Region"
		fmt.Fprintln(res, msg)
		log.Println(msg)
		return
	}

	log.Println("⏳归档……")
	err := archiveOldFiles(region.WorkDir)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		log.Println(err)
		return
	}
	log.Println("✔️归档完成")

	dirPath := path.Join(appConfig.Ftp.Path, "Game")
	name, err := downloadFromFtp(dirPath, region)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		log.Println(err)
		return
	}

	log.Printf("⏳正在解压📦'%s'……\n", name)
	rar := archiver.Rar{
		OverwriteExisting: true,
	}
	err = rar.Unarchive(name, region.WorkDir)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		log.Println(err)
		return
	}
	log.Println("✔️解压完成📦")

	fmt.Fprintln(res, "✔️")
}

func updateConfigHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	lock.Lock()
	defer lock.Unlock()

	_, region := getReqBodyAndRegion(res, req)
	if region == (configRegion{}) {
		res.WriteHeader(http.StatusBadRequest)
		msg := "❓未找到对应的 Region"
		fmt.Fprintln(res, msg)
		log.Println(msg)
		return
	}

	archiveOldFiles(region.WorkDir)

	dirPath := path.Join(appConfig.Ftp.Path, "Config")
	name, err := downloadFromFtp(dirPath, region)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		log.Println(err)
		return
	}

	log.Printf("⏳正在解压📦'%s'……\n", name)

	rar := archiver.Rar{
		OverwriteExisting: true,
	}
	err = rar.Unarchive(name, region.WorkDir)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		log.Println(err)
		return
	}
	log.Println("✔️解压完成📦")

	fmt.Fprintln(res, "✔️")
}

func downloadFromFtp(dirPath string, region configRegion) (string, error) {
	ch := make(chan ftpResult)
	go func() {
		defer close(ch)

		host := appConfig.Ftp.Host
		host = strings.TrimPrefix(host, "ftp://")
		host = strings.TrimPrefix(host, "ftps://")

		addr := fmt.Sprintf("%s:%d", host, appConfig.Ftp.Port)

		log.Println("📡正在连接 ftp 服务器……")
		// ftp目前还没有发布新版本，不能使用 ftp.DialWithShutTimeout(5*time.Second)
		conn, err := ftp.Dial(addr)
		if err != nil {
			ch <- ftpResult{
				err: err,
			}
			return
		}
		log.Println("✔️成功与 ftp 服务器建立连接")
		log.Println("⏳正在登录 ftp……")
		err = conn.Login(appConfig.Ftp.User, appConfig.Ftp.Password)
		if err != nil {
			ch <- ftpResult{
				err: err,
			}
			return
		}
		log.Println("✔️登录成功")

		log.Println("⏳正在获取 ftp 目录信息……")
		entries, err := conn.List(dirPath)
		if err != nil {
			ch <- ftpResult{
				err: err,
			}
			return
		}
		log.Println("✔️成功获取 ftp 目录信息")
		if len(entries) == 0 {
			ch <- ftpResult{
				err: errors.New("❌更新失败，未找到更新包。"),
			}
			return
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Name > entries[j].Name
		})

		name := path.Join(dirPath, entries[0].Name)
		log.Printf("⏳正在下载'%s'……\n", name)
		res, err := conn.Retr(name)
		if err != nil {
			ch <- ftpResult{
				err: err,
			}
			return
		}
		log.Println("✔️已成功获得响应")
		defer res.Close()

		log.Println("⏳正在把响应写入到文件")
		buf, err := ioutil.ReadAll(res)
		if err != nil {
			ch <- ftpResult{
				err: err,
			}
			return
		}
		newName := path.Join(region.WorkDir, entries[0].Name)
		ioutil.WriteFile(newName, buf, 0644)
		log.Println("✔️写入完成")

		ch <- ftpResult{
			file: newName,
		}
	}()
	// 如果 ftp 库更新后，就不用自己处理超时了
	select {
	case <-time.After(time.Duration(appConfig.Ftp.Timeout) * time.Millisecond):
		return "", errors.New("❌Ftp超时")
	case fr := <-ch:
		return fr.file, fr.err
	}
}

func dmpHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	strRegionId := req.URL.Query().Get("regionId")
	regionId, err := strconv.Atoi(strRegionId)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		msg := "❌regionId 不合法"
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

func archiveOldFiles(workDir string) error {
	files, err := ioutil.ReadDir(workDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		name := f.Name()
		if strings.HasSuffix(name, ".Game.rar") || strings.HasSuffix(name, ".Config.rar") {
			go func(f fs.FileInfo) {
				oldpath := path.Join(workDir, name)
				newdir := path.Join(workDir, appConfig.Archive)
				if err := os.MkdirAll(newdir, os.ModePerm); err != nil {
					log.Println(err)
					return
				}
				newpath := path.Join(workDir, appConfig.Archive, name)
				if err := os.Rename(oldpath, newpath); err != nil {
					log.Println(err)
				}
			}(f)
		}
	}
	return nil
}
