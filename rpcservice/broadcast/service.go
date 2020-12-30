package broadcast

import (
	"cbsignal/hub"
	"github.com/lexkong/log"
	"net/rpc"
)

type BroadcastService struct {

}

func RegisterBroadcastService() error {
	log.Infof("register rpcservice service %s", BROADCAST_SERVICE)
	s := new(BroadcastService)
	return rpc.RegisterName(BROADCAST_SERVICE, s)
}

func (b *BroadcastService) Join(request JoinLeaveReq, reply *JoinLeaveResp) error  {
	log.Infof("BroadcastService receive %v", request)
	hub.DoRegisterRemoteClient(request.Id, request.Addr)
	return nil
}
