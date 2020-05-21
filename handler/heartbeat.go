package handler

import (
	"cbsignal/client"
	"github.com/lexkong/log"
)

type HeartbeatHandler struct {
	Cli   *client.Client
}

func (s *HeartbeatHandler)Handle() {

	log.Infof("receive heartbeat from %s", s.Cli.PeerId)

}
