package handler

import (
	"cbsignal/client"
	"cbsignal/hub"
)

type RejectedHandler struct {

	Msg   *SignalMsg
	Cli   *client.Client
}

func (s *RejectedHandler)Handle() {
	h := hub.GetInstance()
	//log.Println(s.Msg.To_peer_id)
	_, ok := h.Clients.Load(s.Msg.To_peer_id)        //判断节点是否还在线
	if ok {
		resp := SignalResp{
			Action: "rejected",
			FromPeerId: s.Cli.PeerId,
		}
		hub.SendJsonToClient(s.Msg.To_peer_id, resp, true)
	}
}
