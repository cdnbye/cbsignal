package hub

import (
	"cbsignal/client"
	"encoding/json"
	"github.com/lexkong/log"
	"sync"
	"sync/atomic"
)

var h *Hub
type Hub struct {

	Clients sync.Map

	ClientNum int64            //count of client

	CompressEnable bool
	CompressLevel int
	CompressRatio int
}

func Init(compressEnable bool, compressLevel int, compressRatio int) {
	h = &Hub{
		CompressEnable: compressEnable,
		CompressLevel: compressLevel,
		CompressRatio: compressRatio,
	}
}

func GetInstance() *Hub {
	return h
}

func DoRegister(client *client.Client) {
	log.Infof("hub DoRegister %s", client.PeerId)
	if client.PeerId != "" {
		h.Clients.Store(client.PeerId, client)
		atomic.AddInt64(&h.ClientNum, 1)
	}
}

func DoRegisterRemoteClient(peerId string, addr string) {
	c := &client.Client{
		LocalNode:    false,
		Conn:         nil,
		PeerId:       peerId,
		InvalidPeers: make(map[string]bool),      // TODO
		RpcNodeAddr:  addr,
	}
	DoRegister(c)
}

func GetClient(peerId string) (*client.Client, bool) {
	cli, ok := h.Clients.Load(peerId)
	if !ok {
		return nil, false
	}
	return cli.(*client.Client), true
}

func DoUnregister(peerId string) {
	log.Infof("hub DoUnregister %s", peerId)
	if peerId == "" {
		return
	}
	_, ok := h.Clients.Load(peerId)
	if ok {
		h.Clients.Delete(peerId)
		atomic.AddInt64(&h.ClientNum, -1)
	}
}

// send json object to a client with peerId
func SendJsonToClient(peerId string, value interface{}, allowCompress bool) {
	b, err := json.Marshal(value)
	if err != nil {
		log.Error("json.Marshal", err)
		return
	}
	cli, ok := h.Clients.Load(peerId)
	if !ok {
		//log.Printf("sendJsonToClient error")
		return
	}
	peer := cli.(*client.Client)
	defer func() {                            // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			log.Warnf(err.(string))                  // 这里的err其实就是panic传入的内容
		}
	}()

	// 小于70的字符串不压缩  TODO
	if h.CompressEnable && allowCompress && peer.CompressSupported && len(b)>=70 {


	} else {

		if err := peer.SendMessage(b); err != nil {
			log.Warnf("sendMessage", err)
		}
	}
}

func GetClientNum() int64 {
	return h.ClientNum
}

func ClearAll()  {
	h.Clients.Range(func(key, value interface{}) bool {
		h.Clients.Delete(key)
		return true
	})
}




