package handler

import "cbsignal/client"

type ExceptionHandler struct {

	Msg      *SignalMsg
	Cli   *client.Client
}

func (s *ExceptionHandler)Handle() {

}