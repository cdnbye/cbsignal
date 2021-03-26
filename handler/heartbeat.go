package handler

import (
	"cbsignal/client"
	"cbsignal/hub"
	"github.com/lexkong/log"
)

type HeartbeatHandler struct {
	Cli   *client.Client
}

func (s *HeartbeatHandler)Handle() {

	log.Infof("receive heartbeat from %s", s.Cli.PeerId)
	s.Cli.UpdateTs()

	resp := SignalResp{
		Action: "pong",
	}
	hub.SendJsonToClient(s.Cli, resp)
}
