package handler

import (
	"cbsignal/client"
	"encoding/json"
)

type Handler interface {
	Handle()
}

type SignalMsg struct {
	Action   string      `json:"action"`
	ToPeerId string      `json:"to_peer_id"`
	Data     interface{} `json:"data"`
}

type SignalResp struct {
	Action string              `json:"action"`
	FromPeerId string          `json:"from_peer_id,omitempty"`
	Data interface{}           `json:"data,omitempty"`
}

func NewHandler(message []byte, cli *client.Client) (Handler, error) {
	signal := SignalMsg{}
	if err := json.Unmarshal(message, &signal); err != nil {
		//log.Println(err)
		return nil, err
	}
	return NewHandlerMsg(signal, cli)
}

func NewHandlerMsg(signal SignalMsg, cli *client.Client) (Handler, error) {
	switch signal.Action {
	case "signal":
		return &SignalHandler{Msg: &signal, Cli: cli}, nil
	case "heartbeat":
		return &HeartbeatHandler{Cli: cli}, nil
	case "rejected":
		return &RejectedHandler{Msg: &signal, Cli: cli}, nil
	default:
		return &ExceptionHandler{Msg: &signal, Cli: cli}, nil
	}
}
