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
	//h := hub.GetInstance()
	//log.Infof("load client Msg %v", s.Msg)
	//判断节点是否还在线
	if target, ok := hub.GetClient(s.Msg.ToPeerId); ok {
		//log.Infof("found client %s", s.Msg.ToPeerId)
		resp := SignalResp{
			Action: "signal",
			FromPeerId: s.Cli.PeerId,
			Data: s.Msg.Data,
		}
		if err := hub.SendJsonToClient(target, resp); err != nil {
			log.Warnf("Send signal to peer %s error %s", target.PeerId, err)
			hub.RemoveClient(target.PeerId)
			s.Cli.EnqueueNotFoundPeer(target.PeerId)
			notFounResp := SignalResp{
				Action: "signal",
				FromPeerId: s.Msg.ToPeerId,
			}
			hub.SendJsonToClient(s.Cli, notFounResp)
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
		//hub.SendJsonToClient(s.Cli.PeerId, resp)
		// 发送一次后，同一peerId下次不再发送，节省sysCall
		if !s.Cli.HasNotFoundPeer(s.Msg.ToPeerId) {
			s.Cli.EnqueueNotFoundPeer(s.Msg.ToPeerId)
			hub.SendJsonToClient(s.Cli, resp)
		}
	}
}


