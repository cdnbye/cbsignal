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
	_, ok := h.Clients.Load(s.Msg.ToPeerId) //判断节点是否还在线
	if ok {
		resp := SignalResp{
			Action: "rejected",
			FromPeerId: s.Cli.PeerId,
			Reason: s.Msg.Reason,
		}
		hub.SendJsonToClient(s.Msg.ToPeerId, resp)
	}
}
