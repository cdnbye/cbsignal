package main
//
//import (
//	"fmt"
//	"github.com/lexkong/log"
//	"github.com/spf13/pflag"
//	"github.com/spf13/viper"
//	"net/http"
//	"os"
//	"os/signal"
//	"strings"
//	"sync/atomic"
//)
//
//var (
//	cfg = pflag.StringP("config", "c", "", "Config file path.")
//)
//
//func init()  {
//	pflag.Parse()
//
//	// Initialize viper
//	if *cfg != "" {
//		viper.SetConfigFile(*cfg) // 如果指定了配置文件，则解析指定的配置文件
//	} else {
//		viper.AddConfigPath("./") // 如果没有指定配置文件，则解析默认的配置文件
//		viper.SetConfigName("config")
//	}
//	viper.SetConfigType("yaml")     // 设置配置文件格式为YAML
//	viper.AutomaticEnv()            // 读取匹配的环境变量
//	replacer := strings.NewReplacer(".", "_")
//	viper.SetEnvKeyReplacer(replacer)
//	if err := viper.ReadInConfig(); err != nil { // viper解析配置文件
//		log.Fatal("Initialize viper", err)
//	}
//
//	// Initialize logger
//	passLagerCfg := log.PassLagerCfg{
//		Writers:        viper.GetString("log.writers"),
//		LoggerLevel:    viper.GetString("log.logger_level"),
//		LoggerFile:     viper.GetString("log.logger_dir"),
//		LogFormatText:  viper.GetBool("log.log_format_text"),
//		RollingPolicy:  viper.GetString("log.rollingPolicy"),
//		LogRotateDate:  viper.GetInt("log.log_rotate_date"),
//		LogRotateSize:  viper.GetInt("log.log_rotate_size"),
//		LogBackupCount: viper.GetInt("log.log_backup_count"),
//	}
//	if err := log.InitWithConfig(&passLagerCfg); err != nil {
//		fmt.Errorf("Initialize logger %s", err)
//	}
//}
//
//var hub *Hub
//
//func main() {
//
//	// Catch SIGINT signals
//	intrChan := make(chan os.Signal)
//	signal.Notify(intrChan, os.Interrupt)
//
//	pflag.Parse()
//
//	SignalPort := viper.GetString("port")
//	SignalPortTLS := viper.GetString("tls.port")
//	signalCert := viper.GetString("tls.cert")
//	signalKey := viper.GetString("tls.key")
//
//	// Initialize viper
//	if *cfg != "" {
//		viper.SetConfigFile(*cfg) // 如果指定了配置文件，则解析指定的配置文件
//	} else {
//		viper.AddConfigPath("./") // 如果没有指定配置文件，则解析默认的配置文件
//		viper.SetConfigName("config")
//	}
//	viper.SetConfigType("yaml")     // 设置配置文件格式为YAML
//	viper.AutomaticEnv()            // 读取匹配的环境变量
//	replacer := strings.NewReplacer(".", "_")
//	viper.SetEnvKeyReplacer(replacer)
//	if err := viper.ReadInConfig(); err != nil { // viper解析配置文件
//		log.Fatal("Initialize viper", err)
//	}
//
//	hub = newHub()
//	go hub.run()
//	http.HandleFunc("/ws", wsHandler)
//	http.HandleFunc("/wss", wsHandler)
//	http.HandleFunc("/", wsHandler)
//
//	http.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
//		//fmt.Printf("URL: %s\n", r.URL.String())
//		w.Header().Set("Access-Control-Allow-Origin", "*")
//		w.Write([]byte(fmt.Sprintf("%d", atomic.LoadInt64(&hub.ClientNum))))
//
//	})
//
//	if  SignalPortTLS != "" && Exists(signalCert) && Exists(signalKey) {
//		go func() {
//			log.Infof("Start to listening the incoming requests on https address: %s\n", SignalPortTLS)
//			err := http.ListenAndServeTLS(SignalPortTLS, signalCert, signalKey, nil)
//			if err != nil {
//				log.Fatal("ListenAndServe: ", err)
//			}
//		}()
//	}
//
//	if SignalPort != "" {
//		go func() {
//			log.Infof("Start to listening the incoming requests on http address: %s\n", SignalPort)
//			err := http.ListenAndServe(SignalPort, nil)
//			if err != nil {
//				log.Fatal("ListenAndServe: ", err)
//			}
//		}()
//	}
//
//	<-intrChan
//
//	log.Info("Shutting down server...")
//}
//
//func wsHandler(w http.ResponseWriter, r *http.Request) {
//	//fmt.Printf("URL: %s\n", r.URL.String())
//	r.ParseForm()
//
//	defer func() {                            // 必须要先声明defer，否则不能捕获到panic异常
//		if err := recover(); err != nil {
//			log.Infof(err.(string))                  // 这里的err其实就是panic传入的内容
//		}
//	}()
//
//	id := r.Form.Get("id")
//	if id != "" {
//		serveWs(hub, w, r, id)
//	}
//}
//
//// 判断所给路径文件/文件夹是否存在
//func Exists(path string) bool {
//	_, err := os.Stat(path)    //os.Stat获取文件信息
//	if err != nil {
//		if os.IsExist(err) {
//			return true
//		}
//		return false
//	}
//	return true
//}
