package rpcservice

type Node interface {

	Init(master bool, addr string)

	IsMaster() bool

	Addr() string

	sendMsgClientJoin()

	sendMsgClientLeft()

	sendMsgClientSignal()

	sendMsgClientRejected()

	Delete()
}
