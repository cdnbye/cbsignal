package handler

import (
	"cbsignal/client"
	"cbsignal/hub"
	"github.com/lexkong/log"
)

type SignalHandler struct {
	Msg   *SignalMsg
	Cli   *client.Client
}

func (s *SignalHandler)Handle() {
	h := hub.GetInstance()
	//log.Infof("load client Msg %v", s.Msg)
	//判断节点是否还在线
	if h.Clients.Has(s.Msg.ToPeerId) {
		//log.Infof("found client %s", s.Msg.ToPeerId)
		resp := SignalResp{
			Action: "signal",
			FromPeerId: s.Cli.PeerId,
			Data: s.Msg.Data,
		}
		if err := hub.SendJsonToClient(s.Msg.ToPeerId, resp); err != nil {
			log.Warnf("Send signal to peer %s error %s", s.Msg.ToPeerId, err)
			//notFounResp := SignalResp{
			//	Action: "signal",
			//	FromPeerId: s.Msg.ToPeerId,
			//}
			//hub.SendJsonToClient(s.Cli.PeerId, notFounResp)
		}
		//if !target.(*client.Client).LocalNode {
		//	log.Warnf("send signal msg from %s to %s on node %s", s.Cli.PeerId, s.Msg.ToPeerId, target.(*client.Client).RpcNodeAddr)
		//}
	} else {
		log.Infof("Peer %s not found", s.Msg.ToPeerId)
		resp := SignalResp{
			Action: "signal",
			FromPeerId: s.Msg.ToPeerId,
		}
		hub.SendJsonToClient(s.Cli.PeerId, resp)
	}
}


