package hub

import (
	"bytes"
	"cbsignal/client"
	"compress/zlib"
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
	//	logrus.Debugf("[Hub.doRegister] %s", client.id)
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
func SendJsonToClient(peerId string, value interface{}, allowCompress bool)  {
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
		var buf bytes.Buffer
		compressor, err := zlib.NewWriterLevel(&buf, h.CompressLevel)
		if err != nil {
			log.Warnf("compress failed %s", err)
			return
		}
		if _, err := compressor.Write(b); err != nil {
			log.Warnf("compress Write failed %s", err)
			return
		}
		if err :=compressor.Close(); err != nil {
			log.Warnf("compress Close failed %s", err)
			return
		}

		log.Infof("before compress len %d", len(b))
		log.Infof("after compress len %d", len(buf.Bytes()))

		if err := peer.SendBinaryData(buf.Bytes()); err != nil {
			log.Warnf("SendBinaryData", err)
		}


	} else {

		if err := peer.SendMessage(b); err != nil {
			log.Warnf("sendMessage", err)
		}
	}
}

func GetClientNum() int64 {
	return h.ClientNum
}




