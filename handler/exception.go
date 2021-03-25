package handler

import (
	"cbsignal/client"
)

type ExceptionHandler struct {

	Msg      *SignalMsg
	Cli   *client.Client
}

// handle {}
func (s *ExceptionHandler)Handle() {
	s.Cli.UpdateTs()
}