package handler

import (
	"cbsignal/client"
	"cbsignal/hub"
)

type SignalHandler struct {

	Msg   *SignalMsg
	Cli   *client.Client
}

func (s *SignalHandler)Handle() {
	h := hub.GetInstance()
	_, ok := h.Clients.Load(s.Msg.To_peer_id)        //判断节点是否还在线
	if ok {
		resp := SignalResp{
			Action: "signal",
			FromPeerId: s.Cli.PeerId,
			Data: s.Msg.Data,
		}

		hub.SendJsonToClient(s.Msg.To_peer_id, resp, true)
	} else {
		//log.Println("Peer not found")
		resp := SignalResp{
			Action: "signal",
			FromPeerId: s.Msg.To_peer_id,
		}
		// 发送一次后，同一peerId下次不再发送，节省带宽
		if !s.Cli.InvalidPeers[s.Msg.To_peer_id] {
			s.Cli.InvalidPeers[s.Msg.To_peer_id] = true
			hub.SendJsonToClient(s.Cli.PeerId, resp, true)
		}
	}
}


