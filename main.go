package main

import (
	"bytes"
	"cbsignal/client"
	"cbsignal/handler"
	"cbsignal/hub"
	"cbsignal/rpcservice/broadcast"
	"cbsignal/rpcservice/heartbeat"
	signalservice "cbsignal/rpcservice/signalmsg"
	"cbsignal/util"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/lexkong/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	cfg = pflag.StringP("config", "c", "", "Config file path.")
	flagMaster = pflag.BoolP("master", "m", false, "Master or slave")
	newline = []byte{'\n'}
	space   = []byte{' '}

	allowMap = make(map[string]bool)            // allow list of domain
	useAllowList = false
	blockMap = make(map[string]bool)            // block list of domain
	useBlockList = false

	isMaster bool
	masterIp string
	masterPort string
	selfIp string
	selfPort string

	signalPort     string
	signalPortTLS  string
	signalCertPath string
	signalKeyPath  string

	version string
	compressionEnabled bool
	compressionLevel int
	compressionActivationRatio int

	broadcastClient *broadcast.Client
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

	selfIp = util.GetInternalIP()
	selfPort = viper.GetString("cluster.self_port")
	masterIp = viper.GetString("cluster.master.ip")
	if masterIp == "127.0.0.1" || masterIp == "localhost" || masterIp == "0.0.0.0" {
		isMaster = true
	} else if masterIp == "" {
		//masterIp = "127.0.0.1"
		selfIp = "127.0.0.1"
	}
	masterPort = viper.GetString("cluster.master.port")

	//isMaster = selfIp == masterIp
	if *flagMaster {
		isMaster = true
	}

	log.Warnf("isMaster %t", isMaster)

	signalPort = viper.GetString("port")
	signalPortTLS = viper.GetString("tls.port")
	signalCertPath = viper.GetString("tls.cert")
	signalKeyPath = viper.GetString("tls.key")

	version = viper.GetString("version")
	compressionEnabled = viper.GetBool("compression.enable")
	compressionLevel = viper.GetInt("compression.level")
	compressionActivationRatio = viper.GetInt("compression.activationRatio")

}

func main() {

	// Catch SIGINT signals
	intrChan := make(chan os.Signal)
	signal.Notify(intrChan, os.Interrupt)

	// Increase resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	hub.Init(compressionEnabled, compressionLevel, compressionActivationRatio)

	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/wss", wsHandler)
	http.HandleFunc("/", wsHandler)
	http.HandleFunc("/count", handler.CountHandler())
	http.HandleFunc("/version", handler.VersionHandler(version))
	http.HandleFunc("/info", handler.StatsHandler(version, compressionEnabled))

	// rpcservice
	go func() {
		// 注册rpc心跳服务
		log.Warnf("register rpcservice service on tcp address: %s\n", selfPort)
		if err := heartbeat.RegisterHeartbeatService();err != nil {
			panic(err)
		}
		listener, err := net.Listen("tcp", selfPort)
		if err != nil {
			log.Fatal("ListenTCP error:", err)
		}
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal("Accept error:", err)
			}
			go rpc.ServeConn(conn)
		}
	}()
	time.Sleep(2*time.Second)

	masterAddr := masterIp + masterPort
	selfAddr := selfIp + selfPort
	log.Infof("DialHeartbeatService %s", masterAddr)
	heartbeatClient := heartbeat.NewHeartbeatClient(masterAddr, selfAddr)
	heartbeatClient.DialHeartbeatService()
	heartbeatClient.StartHeartbeat()

	broadcastClient = broadcast.NewBroadcastClient(heartbeatClient.NodeHub(), selfAddr)
	// 注册rpc广播服务
	if err := broadcast.RegisterBroadcastService();err != nil {
		panic(err)
	}

	// 注册rpc信令服务
	if err := signalservice.RegisterSignalService();err != nil {
		panic(err)
	}

	if  signalPortTLS != "" && util.Exists(signalCertPath) && util.Exists(signalKeyPath) {
		go func() {
			log.Warnf("Start to listening the incoming requests on https address: %s\n", signalPortTLS)
			err := http.ListenAndServeTLS(signalPortTLS, signalCertPath, signalKeyPath, nil)
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
		}()
	}

	if signalPort != "" {
		go func() {
			log.Warnf("Start to listening the incoming requests on http address: %s\n", signalPort)
			err := http.ListenAndServe(signalPort, nil)
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
		}()
	}

	<-intrChan

	log.Info("Shutting down server...")
}



func wsHandler(w http.ResponseWriter, r *http.Request) {

	// Upgrade connection
	//log.Printf("UpgradeHTTP")
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}

	origin := r.Header.Get("Origin")
	if origin != "" {
		domain := util.GetDomain(origin)
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
		LocalNode:    true,
		Conn:         conn,
		PeerId:       id,
		InvalidPeers: make(map[string]bool),
	}
	hub.DoRegister(c)
	broadcastClient.BroadcastMsgJoin(id)

	go func() {
		defer func() {
			// 节点离开
			log.Infof("peer leave")
			hub.DoUnregister(id)
			conn.Close()
			broadcastClient.BroadcastMsgLeave(id)
		}()

		for {
			msg, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				// handle error
				//log.Printf("ReadClientData " + err.Error())
				break
			}
			//log.Infof("ReadClientData from " + id)
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


