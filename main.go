package main

import (
	"bytes"
	"cbsignal/client"
	"cbsignal/handler"
	"cbsignal/hub"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/lexkong/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	cfg = pflag.StringP("config", "c", "", "Config file path.")
	newline = []byte{'\n'}
	space   = []byte{' '}
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
}

//var EventsCollector *eventsCollector

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	//log.Printf("UpgradeHTTP")
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
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
			log.Infof("节点离开1 ")
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

	hub.Init()

	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/wss", wsHandler)
	http.HandleFunc("/", wsHandler)
	http.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("URL: %s\n", r.URL.String())
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(fmt.Sprintf("%d", hub.GetClientNum())))

	})

	if  SignalPortTLS != "" && Exists(signalCert) && Exists(signalKey) {
		go func() {
			log.Infof("Start to listening the incoming requests on https address: %s\n", SignalPortTLS)
			err := http.ListenAndServeTLS(SignalPortTLS, signalCert, signalKey, nil)
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
		}()
	}

	if SignalPort != "" {
		go func() {
			log.Infof("Start to listening the incoming requests on http address: %s\n", SignalPort)
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