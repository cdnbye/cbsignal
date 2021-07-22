package handler

import (
	"cbsignal/hub"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

type SignalInfo struct {
	Version string `json:"version"`
	CurrentConnections int `json:"current_connections"`
	RateLimit          int64  `json:"rate_limit,omitempty"`
	SecurityEnabled    bool `json:"security_enabled,omitempty"`
	ClusterMode        bool `json:"cluster_mode,omitempty"`
	NumGoroutine       int  `json:"num_goroutine"`
	NumPerMap          []int `json:"num_per_map"`
}

type Resp struct {
	Ret int `json:"ret"`
	Data *SignalInfo `json:"data"`
}

func StatsHandler(info SignalInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("URL: %s\n", r.URL.String())
		info.NumGoroutine = runtime.NumGoroutine()
		info.NumPerMap = hub.GetClientNumPerMap()
		info.CurrentConnections = 0
		for _, count := range info.NumPerMap {
			info.CurrentConnections += count
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		resp := Resp{
			Ret:  0,
			Data: &info,
		}
		b, err := json.MarshalIndent(resp, "", "   ")
		if err != nil {
			resp, _ := json.Marshal(Resp{
				Ret:  -1,
				Data: nil,
			})
			w.Write(resp)
			return
		}
		w.Write(b)
	}
}

func VersionHandler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("URL: %s\n", r.URL.String())
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(fmt.Sprintf("%s", version)))

	}
}

func CountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("URL: %s\n", r.URL.String())
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(fmt.Sprintf("%d", hub.GetClientNum())))

	}
}

