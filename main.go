package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"io/ioutil"
	"errors"
	"strconv"
	"strings"
	"path/filepath"
	"os/user"
	"log"
)

type HostConfig struct {
	Hosts []Host
	Localhost string
	Port int
}

type Host struct {
	Hostname string
	Basedir string
}


func check(e error) {
	if e != nil {
		panic(e)
	}
}

func muxHost(router *mux.Router, host Host) {
	path := host.Basedir
	if strings.HasPrefix(path, "~/") {
		usr, _ := user.Current()
		dir := usr.HomeDir
		path = filepath.Join(dir, path[2:])
	}
	handler := http.FileServer(http.Dir(path))
	if host.Hostname == "/" {
		router.PathPrefix("/").Handler(handler)
	}else {
		router.Host(host.Hostname).PathPrefix("/").Handler(handler)
		println("Host added: " + host.Hostname + " -> " + path)
	}
}

func getLocalhost(config HostConfig) Host {
	for _, host := range config.Hosts {
		if host.Hostname == config.Localhost {
			println("Default Host Set: "+host.Hostname)
			return Host{"/", host.Basedir}
		}
	}
	panic(errors.New("localhost specified in config is not in hosts list"))
}

func main()  {
	r := mux.NewRouter()

	var hostConfig HostConfig
	dat, err := ioutil.ReadFile(".host-config.json")
	check(err)
	json.Unmarshal(dat, &hostConfig)

	for _, host := range hostConfig.Hosts {
		muxHost(r, host)
	}

	localhost := getLocalhost(hostConfig)
	muxHost(r, localhost)

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(hostConfig.Port), r))
}


