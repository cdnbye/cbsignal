package hub

import (
	"cbsignal/client"
	"cbsignal/util/cmap"
	"encoding/json"
	"fmt"
	"github.com/lexkong/log"
)

var h *Hub
type Hub struct {

	Clients cmap.ConcurrentMap

	CompressEnable bool
	CompressLevel int
	CompressRatio int
}

func Init(compressEnable bool, compressLevel int, compressRatio int) {
	h = &Hub{
		Clients: cmap.New(),
		CompressEnable: compressEnable,
		CompressLevel: compressLevel,
		CompressRatio: compressRatio,
	}

}

func GetInstance() *Hub {
	return h
}

func GetClientNum() int {
	return h.Clients.CountNoLock()
}

func GetClientNumPerMap() []int {
	return h.Clients.CountPerMapNoLock()
}

func DoRegister(client *client.Client) {
	log.Infof("hub DoRegister %s", client.PeerId)
	if client.PeerId != "" {
		h.Clients.Set(client.PeerId, client)
	}
}

func DoRegisterRemoteClient(peerId string, addr string) {
	if peerId == "" || addr == "" {
		log.Warnf("Invalid peer %s from addr %s", peerId, addr)
		return
	}
	c := client.NewPeerClient(peerId, nil, false, addr)
	DoRegister(c)
}

func GetClient(peerId string) (*client.Client, bool) {
	cli, ok := h.Clients.Get(peerId)
	if !ok {
		return nil, false
	}
	return cli.(*client.Client), true
}

func DoUnregister(peerId string) bool {
	log.Infof("hub DoUnregister %s", peerId)
	if peerId == "" {
		return false
	}
	if h.Clients.Has(peerId) {
		h.Clients.Remove(peerId)
		return true
	}
	return false
}

// send json object to a client with peerId
func SendJsonToClient(peerId string, value interface{}) error {
	b, err := json.Marshal(value)
	if err != nil {
		log.Error("json.Marshal", err)
		return err
	}
	cli, ok := h.Clients.Get(peerId)
	if !ok {
		//log.Printf("sendJsonToClient error")
		return fmt.Errorf("peer %s not found", peerId)
	}
	peer := cli.(*client.Client)
	defer func() {                            // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			log.Warnf(err.(string))                  // 这里的err其实就是panic传入的内容
		}
	}()

	// 如果开启压缩  TODO
	if h.CompressEnable {


	} else {
		if err := peer.SendMessage(b); err != nil {
			//log.Warnf("sendMessage", err)
			return err
		}
	}
	return nil
}

func ClearAll()  {
	h.Clients.Clear()
}




