package main

import (
	"bytes"
	"cbsignal/client"
	"cbsignal/handler"
	"cbsignal/hub"
	"cbsignal/rpcservice/broadcast"
	"cbsignal/rpcservice/heartbeat"
	"cbsignal/rpcservice/signaling"
	"cbsignal/util"
	"cbsignal/util/ratelimit"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/lexkong/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"net"
	"net/http"
	//_ "net/http/pprof"
	"net/rpc"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	VERSION = "2.2.0"
	CHECK_CLIENT_INTERVAL = 15 * 60
	EXPIRE_LIMIT = 15 * 60
)

var (
	cfg = pflag.StringP("config", "c", "", "Config file path.")
	newline = []byte{'\n'}
	space   = []byte{' '}

	masterIp string
	masterPort string
	selfIp string
	selfPort string
	isCluster bool
	masterAddr string
	selfAddr string

	signalPort     string
	signalPortTLS  string
	signalCertPath string
	signalKeyPath  string

	versionNum int
	compressionEnabled bool
	compressionLevel int
	compressionActivationRatio int

	limitEnabled bool
	limitRate    int64
	limiter      *ratelimit.Bucket

	broadcastClient *broadcast.Client
	heartbeatClient *heartbeat.Client

	securityEnabled bool
	maxTimeStampAge int64
	securityToken string
)

func init()  {
	pflag.Parse()

	versionNum = util.GetVersionNum(VERSION)

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

	isCluster = viper.GetBool("cluster.enable")
	if isCluster {
		selfIp = viper.GetString("cluster.self.ip")
		if selfIp == "" {
			selfIp = util.GetInternalIP()
		}
		selfPort = viper.GetString("cluster.self.port")
		masterIp = viper.GetString("cluster.master.ip")
		if masterIp == "" {
			// master node
			masterIp = selfIp
		} else if masterIp == "127.0.0.1" || masterIp == "localhost" || masterIp == "0.0.0.0" {
			// cluster in local
			masterIp = ""
			selfIp = ""
		}
		masterPort = viper.GetString("cluster.master.port")
		if masterPort == "" {
			masterPort = selfPort
		}
	}

	signalPort = viper.GetString("port")
	signalPortTLS = viper.GetString("tls.port")
	signalCertPath = viper.GetString("tls.cert")
	signalKeyPath = viper.GetString("tls.key")

	setupConfigFromViper()

	//开始监听
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Warnf("config is changed :%s \n", e.Name)
		setupConfigFromViper()
	})

	hub.Init(compressionEnabled, compressionLevel, compressionActivationRatio)
	go func() {
		for {
			time.Sleep(CHECK_CLIENT_INTERVAL*time.Second)
			now := time.Now().Unix()
			log.Warnf("start check client alive...")
			count := 0
			for item := range hub.GetInstance().Clients.IterBuffered() {
				cli := item.Val
				if cli.LocalNode && cli.IsExpired(now, EXPIRE_LIMIT) {
					// 节点过期
					//log.Warnf("client %s is expired for %d, close it", cli.PeerId, now-cli.Timestamp)
					if ok := hub.DoUnregister(cli.PeerId); ok {
						cli.Close()
						count ++
					}
				}
			}
			log.Warnf("check client finished, closed %d clients", count)
		}
	}()
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

	// rate limit
	if limitEnabled {
		log.Warnf("Init ratelimit with rate %d", limitRate)
		limiter = ratelimit.NewBucketWithQuantum(time.Second, limitRate, limitRate)
	}

	if securityEnabled {
		if maxTimeStampAge == 0 || securityToken == "" {
			panic("maxTimeStampAge or token is empty when security on")
		}
		if len(securityToken) > 8 {
			panic("security token is larger than 8")
		}
		log.Warnf("security on\nmaxTimeStampAge %d\ntoken %s", maxTimeStampAge, securityToken)
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

	if  signalPortTLS != "" && util.Exists(signalCertPath) && util.Exists(signalKeyPath) {
		go func() {
			log.Warnf("Start to listening the incoming requests on https address: %s\n", signalPortTLS)
			err := http.ListenAndServeTLS(signalPortTLS, signalCertPath, signalKeyPath, nil)
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
		}()
	}

	if isCluster {
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
		time.Sleep(6*time.Second)

		masterAddr = masterIp + masterPort
		selfAddr = selfIp + selfPort
		log.Infof("DialHeartbeatService %s", masterAddr)
		heartbeatClient = heartbeat.NewHeartbeatClient(masterAddr, selfAddr)
		heartbeatClient.DialHeartbeatService()
		heartbeatClient.StartHeartbeat()

		broadcastClient = broadcast.NewBroadcastClient(heartbeatClient.NodeHub(), selfAddr)
		// 注册rpc广播服务
		if err := broadcast.RegisterBroadcastService();err != nil {
			panic(err)
		}

		// 注册rpc信令服务
		if err := signaling.RegisterSignalService();err != nil {
			panic(err)
		}
	}

	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/wss", wsHandler)
	http.HandleFunc("/", wsHandler)
	http.HandleFunc("/count", handler.CountHandler())
	http.HandleFunc("/version", handler.VersionHandler(VERSION))

	info := handler.SignalInfo{
		Version:            VERSION,
		CompressionEnabled: compressionEnabled,
		SecurityEnabled: securityEnabled,
		ClusterMode: isCluster,
	}
	if limitEnabled {
		info.RateLimit = limitRate
	}
	http.HandleFunc("/info", handler.StatsHandler(info))

	<-intrChan

	broadcastClient.GetNodeHub().Clear()
	log.Info("Shutting down server...")
}

func wsHandler(w http.ResponseWriter, r *http.Request) {

	// Upgrade connection
	//log.Printf("UpgradeHTTP")
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}

	// 限流
	if limitEnabled {
		if limiter.TakeAvailable(1) == 0 {
			log.Warnf("reach rate limit %d", limiter.Capacity())
			conn.Close()
			return
		} else {
			log.Infof("rate limit remaining %d capacity %d", limiter.Available(), limiter.Capacity())
		}
	}

	r.ParseForm()
	id := r.Form.Get("id")
	//platform := r.Form.Get("p")
	//log.Printf("id %s", id)
	if id == "" {
		conn.Close()
		return
	}

	c := client.NewPeerClient(id, conn, true, selfAddr)

	// 校验
	if securityEnabled {
		now := time.Now().Unix()
		token := strings.Split(r.Form.Get("token"), "-")
		if len(token) < 2 {
			log.Warnf("token not valid %s origin %s", r.Form.Get("token"), r.Header.Get("Origin"))
			closeInvalidConn(c)
			return
		}
		sign := token[0]
		tsStr := token[1]
		if ts, err := strconv.ParseInt(tsStr, 10, 64); err != nil {
			//log.Warnf("ts ParseInt", err)
			closeInvalidConn(c)
			return
		} else {
			if ts < now - maxTimeStampAge || ts > now + maxTimeStampAge  {
				log.Warnf("ts expired for %d origin %s", now - ts, r.Header.Get("Origin"))
				closeInvalidConn(c)
				return
			}
			//ip := strings.Split(r.RemoteAddr, ":")[0]
			//log.Infof("client ip %s", ip)

			hm := hmac.New(md5.New, []byte(securityToken))
			hm.Write([]byte(tsStr + id))
			realSign := hex.EncodeToString(hm.Sum(nil))[:8]
			if sign != realSign {
				log.Warnf("client token %s not match %s origin %s", sign, realSign, r.Header.Get("Origin"))
				closeInvalidConn(c)
				return
			}
		}
	}

	hub.DoRegister(c)
	if isCluster {
		go broadcastClient.BroadcastMsgJoin(id)
	}
	// 发送版本号
	c.SendMsgVersion(versionNum)

	go func() {
		defer func() {
			// 节点离开
			log.Infof("peer leave")
			if ok := hub.DoUnregister(id); ok {
				//log.Infof("close peer %s", id)
				closeInvalidConn(c)
			}
			if isCluster {
				go broadcastClient.BroadcastMsgLeave(id)
			}
		}()
		msg := make([]wsutil.Message, 0, 4)
		for {
			msg, err = wsutil.ReadClientMessage(conn, msg[:0])
			if err != nil {
				log.Infof("read message error: %v", err)
				break
			}
			for _, m := range msg {
				// ping
				if m.OpCode.IsControl() {
					c.UpdateTs()
					//log.Warnf("receive ping from %s platform %s", id, platform)
					err := wsutil.HandleClientControlMessage(conn, m)
					if err != nil {
						log.Infof("handle control error: %v", err)
					}
					continue
				}
				data := bytes.TrimSpace(bytes.Replace(m.Payload, newline, space, -1))
				hdr, err := handler.NewHandler(data, c)
				if err != nil {
					// 心跳包
					c.UpdateTs()
					log.Infof("NewHandler " + err.Error())
				} else {
					hdr.Handle()
				}
			}
		}
	}()
}

func closeInvalidConn(cli *client.Client)  {
	cli.SendMsgClose("invalid client")
	cli.Close()
}

func setupConfigFromViper()  {
	compressionEnabled = viper.GetBool("compression.enable")
	compressionLevel = viper.GetInt("compression.level")
	compressionActivationRatio = viper.GetInt("compression.activationRatio")
	limitEnabled = viper.GetBool("ratelimit.enable")
	limitRate = viper.GetInt64("ratelimit.max_rate")
	securityEnabled = viper.GetBool("security.enable")
	maxTimeStampAge = viper.GetInt64("security.maxTimeStampAge")
	securityToken = viper.GetString("security.token")
}


