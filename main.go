package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	os "os"
)

type HostConfig struct {
	Hosts     []Host
	Localhost string
	Port      int
}

type Host struct {
	Hostname string
	Basedir  string
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
	} else {
		router.Host(host.Hostname).PathPrefix("/").Handler(handler)
		println("Host added: " + host.Hostname + " -> " + path)
	}
}

func getLocalhost(config HostConfig) Host {
	for _, host := range config.Hosts {
		if host.Hostname == config.Localhost {
			println("Default Host Set: " + host.Hostname)
			return Host{"/", host.Basedir}
		}
	}
	panic(errors.New("localhost specified in config is not in hosts list"))
}

func main() {
	r := mux.NewRouter()

	var hostConfig HostConfig
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	configPath := filepath.Join(dir, ".host-config.json")
	dat, err := ioutil.ReadFile(configPath)
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
