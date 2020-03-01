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

}

func Init() {
	h = &Hub{

	}
}

func GetInstance() *Hub {
	return h
}

func DoRegister(client *client.Client) {
	//	logrus.Debugf("[Hub.doRegister] %s", client.id)
	if client.PeerId != "" {
		h.Clients.Store(client.PeerId, client)
		atomic.AddInt64(&h.ClientNum, 1)
	}
}

func DoUnregister(client *client.Client) {
	//	logrus.Debugf("[Hub.doUnregister] %s", client.id)

	if client.PeerId == "" {
		return
	}
	atomic.AddInt64(&h.ClientNum, -1)
	_, ok := h.Clients.Load(client.PeerId)
	if ok {
		h.Clients.Delete(client.PeerId)
	}

}

// send json object to a client with peerId
func SendJsonToClient(peerId string, value interface{})  {
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
	defer func() {                            // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			log.Warnf(err.(string))                  // 这里的err其实就是panic传入的内容
		}
	}()
	if err := cli.(*client.Client).SendMessage(b); err != nil {
		log.Warnf("sendMessage", err)
	}
}

func GetClientNum() int64 {
	return h.ClientNum
}




