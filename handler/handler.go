package handler

type Handler interface {
	Handle()
}

type SignalMsg struct {
	To_peer_id string          `json:"to_peer_id"`
	Data  interface{}          `json:"data"`
}

type SignalResp struct {
	Action string              `json:"action"`
	FromPeerId string          `json:"from_peer_id"`
	Data interface{}           `json:"data,omitempty"`
}
