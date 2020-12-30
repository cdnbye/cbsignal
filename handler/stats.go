package handler

import (
	"cbsignal/hub"
	"encoding/json"
	"fmt"
	"net/http"
)

type SignalInfo struct {
	Version string `json:"version"`
	CurrentConnections int64 `json:"current_connections"`
	CompressionEnabled bool `json:"compression_enabled"`
}

type Resp struct {
	Ret int `json:"ret"`
	Data *SignalInfo `json:"data"`
}

func StatsHandler(version string, compressionEnabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("URL: %s\n", r.URL.String())
		w.Header().Set("Access-Control-Allow-Origin", "*")
		info := SignalInfo{
			Version:            version,
			CurrentConnections: hub.GetClientNum(),
			CompressionEnabled: compressionEnabled,
		}
		resp := Resp{
			Ret:  0,
			Data: &info,
		}
		b, err := json.Marshal(resp)
		if err != nil {
			resp, _ := json.Marshal(Resp{
				Ret:  -1,
				Data: nil,
			})
			w.Write(resp)
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

