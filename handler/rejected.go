package handler

import (
	"cbsignal/client"
	"cbsignal/hub"
)

type RejectHandler struct {
	Msg   *SignalMsg
	Cli   *client.Client
}

func (s *RejectHandler)Handle() {
	h := hub.GetInstance()
	_, ok := h.Clients.Load(s.Msg.ToPeerId) //判断节点是否还在线
	if ok {
		resp := SignalResp{
			Action: "reject",
			FromPeerId: s.Cli.PeerId,
			Reason: s.Msg.Reason,
		}
		hub.SendJsonToClient(s.Msg.ToPeerId, resp)
	}
}
