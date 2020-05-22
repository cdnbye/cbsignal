package handler

import (
	"cbsignal/client"
	"cbsignal/hub"
	"github.com/lexkong/log"
	"math/rand"
	"time"
)

type CompressHandler struct {

	Msg   *SignalMsg
	Cli   *client.Client
}

func (s *CompressHandler)Handle() {
	h := hub.GetInstance()

	if h.CompressEnable && s.Msg.Supported && h.CompressRatio > 0 {
		x := 100
		if h.CompressRatio != 100 {
			rand.Seed(time.Now().UnixNano())
			x = rand.Intn(101) //生成0-100随机整数
		}
		log.Infof("rand x: %d CompressRatio: %d", x, h.CompressRatio)
		if x >= h.CompressRatio {
			s.Cli.CompressSupported = true
			resp := SignalResp{
				Action: "compress",
				Supported: true,
			}
			hub.SendJsonToClient(s.Cli.PeerId, resp, false)
		}
	}
}
