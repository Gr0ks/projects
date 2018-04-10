package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Config struct {
	Miners map[string]map[string]*Miner `json:"miners"`
	Listen string                       `json:"listen"`
}

type Server struct {
	Config *Config
}

func getInts(s string) []int64 {
	parts := strings.Split(s, ";")
	var arr []int64
	for _, p := range parts {
		i, _ := strconv.ParseInt(p, 10, 64)
		arr = append(arr, i)
	}
	return arr
}

func LoadConfigFile() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	cfg := Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
}

func (srv *Server) RebootHandler(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)

	ip := mux.Vars(r)["ip"]
	passwd := mux.Vars(r)["passwd"]
	if passwd == "Your passwd" {
		go func() {
			err := Reboot(ip)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				log.Println(err.Error())
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(ip + " Rebooted"))
				log.Println(ip, "reboot")
			}
		}()
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func (srv *Server) MinersHandler(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(srv.Config.Miners)
	if err != nil {
		log.Println("Error serializing /miners: ", err)
	}
}

func (srv *Server) MinerHandler(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)
	user := mux.Vars(r)["user"]
	miner, ok := srv.Config.Miners[user]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	reply := make(map[string]map[string]*Miner)
	reply[user] = miner
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Println("Error serializing /miners/user ", err)
	}
}

func main() {
	cfg, err := LoadConfigFile()
	if err != nil {
		log.Panic(err)
		return
	}

	server := Server{}
	server.Config = cfg

	r := mux.NewRouter()

	r.HandleFunc("/miners", server.MinersHandler)
	r.HandleFunc("/miners/{user}", server.MinerHandler)
	r.HandleFunc("/miners/reboot/{ip}/{passwd}", server.RebootHandler)
	go func() {
		l, err := net.Listen("tcp", server.Config.Listen)
		if err != nil {
			log.Panic(err)
		} else {
			if err := http.Serve(l, r); err != nil {
				log.Panic(err)
			}
		}
	}()

	if err != nil {
		log.Panic(err)
		return
	}

	intv := time.Duration(time.Second * 10)
	timer := time.NewTimer(intv)
	for _, miner := range cfg.Miners {
		for _, worker := range miner {
			go worker.GetStatus()
		}
	}
	for {
		select {
		case <-timer.C:
			for _, miner := range cfg.Miners {
				for _, worker := range miner {
					go worker.GetStatus()
				}
			}
			timer.Reset(intv)
		}
	}
}
