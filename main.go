package main

import (
	"bytes"
	"cbsignal/client"
	"cbsignal/handler"
	"cbsignal/hub"
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/lexkong/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	cfg = pflag.StringP("config", "c", "", "Config file path.")
	newline = []byte{'\n'}
	space   = []byte{' '}

	allowMap = make(map[string]bool)            // allow list of domain
	useAllowList = false
	blockMap = make(map[string]bool)            // block list of domain
	useBlockList = false

	capacity int64 = 30000
)

func init()  {
	pflag.Parse()

	// Initialize viper
	if *cfg != "" {
		viper.SetConfigFile(*cfg) // 如果指定了配置文件，则解析指定的配置文件
	} else {
		viper.AddConfigPath("./") // 如果没有指定配置文件，则解析默认的配置文件
		viper.SetConfigName("config")
	}
	viper.SetConfigType("yaml")     // 设置配置文件格式为YAML
	viper.AutomaticEnv()            // 读取匹配的环境变量
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil { // viper解析配置文件
		log.Fatal("Initialize viper", err)
	}

	// Initialize logger
	passLagerCfg := log.PassLagerCfg{
		Writers:        viper.GetString("log.writers"),
		LoggerLevel:    viper.GetString("log.logger_level"),
		LoggerFile:     viper.GetString("log.logger_dir"),
		LogFormatText:  viper.GetBool("log.log_format_text"),
		RollingPolicy:  viper.GetString("log.rollingPolicy"),
		LogRotateDate:  viper.GetInt("log.log_rotate_date"),
		LogRotateSize:  viper.GetInt("log.log_rotate_size"),
		LogBackupCount: viper.GetInt("log.log_backup_count"),
	}
	if err := log.InitWithConfig(&passLagerCfg); err != nil {
		fmt.Errorf("Initialize logger %s", err)
	}

	// Initialize allow list and block list
	allowList := viper.GetStringSlice("allow_list")
	if len(allowList) > 0 {
		useAllowList = true
		for _, v := range allowList {
			allowMap[v] = true
		}
	}
	blockList := viper.GetStringSlice("block_list")
	if len(blockList) > 0 {
		useBlockList = true
		for _, v := range blockList{
			blockMap[v] = true
		}
	}
	if useBlockList && useAllowList {
		panic("Do not use allowList and blockList at the same time")
	}

    capacity = viper.GetInt64("capacity")
    if capacity <= 0 {
		panic("capacity <= 0")
	}
}

//var EventsCollector *eventsCollector

func wsHandler(w http.ResponseWriter, r *http.Request) {

	// Upgrade connection
	//log.Printf("UpgradeHTTP")
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}

	origin := r.Header.Get("Origin")
	if origin != "" {
		domain := GetDomain(origin)
		log.Debugf("domain: %s", domain)
		if useAllowList && !allowMap[domain] {
			log.Infof("domian %s is out of allowList", domain)
			//wsutil.WriteServerMessage(conn, ws.OpClose, nil)
			conn.Close()
			return
		} else if useBlockList && blockMap[domain] {
			log.Infof("domian %s is in blockList", domain)
			//wsutil.WriteServerMessage(conn, ws.OpClose, nil)
			conn.Close()
			return
		}
	}

	r.ParseForm()
	id := r.Form.Get("id")
	//log.Printf("id %s", id)
	if id == "" {
		conn.Close()
		return
	}

	c := &client.Client{
		Conn: conn,
		PeerId: id,
		InvalidPeers: make(map[string]bool),
	}
	hub.DoRegister(c)

	go func() {
		defer func() {
			// 节点离开
			log.Infof("peer leave")
			hub.DoUnregister(c)
			conn.Close()
		}()

		for {
			msg, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				// handle error
				//log.Printf("ReadClientData " + err.Error())
				break
			}
			//log.Printf("ReadClientData " + string(msg))
			msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
			hdr, err := handler.NewHandler(msg, c)
			if err != nil {
				// 心跳包
				log.Infof("NewHandler " + err.Error())
			} else {
				hdr.Handle()
			}
		}
	}()
}

type Resp struct {
	Ret int `json:"ret"`
	Data *SignalInfo `json:"data"`
} 

type SignalInfo struct {
	Version string `json:"version"`
	CurrentConnections int64 `json:"current_connections"`
	Capacity int64 `json:"capacity"`
	UtilizationRate float32 `json:"utilization_rate"`
	CompressionEnabled bool `json:"compression_enabled"`
}

func main() {

	// Catch SIGINT signals
	intrChan := make(chan os.Signal)
	signal.Notify(intrChan, os.Interrupt)

	pflag.Parse()

	SignalPort := viper.GetString("port")
	SignalPortTLS := viper.GetString("tls.port")
	signalCert := viper.GetString("tls.cert")
	signalKey := viper.GetString("tls.key")

	// Initialize viper
	if *cfg != "" {
		viper.SetConfigFile(*cfg) // 如果指定了配置文件，则解析指定的配置文件
	} else {
		viper.AddConfigPath("./") // 如果没有指定配置文件，则解析默认的配置文件
		viper.SetConfigName("config")
	}
	viper.SetConfigType("yaml")     // 设置配置文件格式为YAML
	viper.AutomaticEnv()            // 读取匹配的环境变量
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil { // viper解析配置文件
		log.Fatal("Initialize viper", err)
	}

	// Increase resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	// Enable pprof hooks
	//go func() {
	//	if err := http.ListenAndServe("localhost:6060", nil); err != nil {
	//		log.Fatalf("pprof failed: %v", err)
	//	}
	//}()

	hub.Init(viper.GetBool("compression.enable"), viper.GetInt("compression.level"), viper.GetInt("compression.activationRatio"))

	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/wss", wsHandler)
	http.HandleFunc("/", wsHandler)
	http.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("URL: %s\n", r.URL.String())
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(fmt.Sprintf("%d", hub.GetClientNum())))

	})
	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("URL: %s\n", r.URL.String())
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(fmt.Sprintf("%s", viper.GetString("version"),)))

	})
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("URL: %s\n", r.URL.String())
		w.Header().Set("Access-Control-Allow-Origin", "*")
		currentConnections := hub.GetClientNum()
		utilizationRate := float32(currentConnections)/float32(capacity)
		info := SignalInfo{
			Version:            viper.GetString("version"),
			CurrentConnections: hub.GetClientNum(),
			Capacity:           capacity,
			UtilizationRate:    utilizationRate,
			CompressionEnabled: viper.GetBool("compression.enable"),
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
	})

	if  SignalPortTLS != "" && Exists(signalCert) && Exists(signalKey) {
		go func() {
			log.Warnf("Start to listening the incoming requests on https address: %s\n", SignalPortTLS)
			err := http.ListenAndServeTLS(SignalPortTLS, signalCert, signalKey, nil)
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
		}()
	}

	if SignalPort != "" {
		go func() {
			log.Warnf("Start to listening the incoming requests on http address: %s\n", SignalPort)
			err := http.ListenAndServe(SignalPort, nil)
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
		}()
	}

	<-intrChan

	log.Info("Shutting down server...")
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)    //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// 获取域名（不包含端口）
func GetDomain(uri string) string {
	parsed, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	a := strings.Split(parsed.Host, ":")
	return a[0]
}