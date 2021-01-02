package broadcast

import (
	"cbsignal/hub"
	"cbsignal/rpcservice"
	"github.com/lexkong/log"
	"net/rpc"
)

type Service struct {

}

func RegisterBroadcastService() error {
	log.Infof("register rpc service %s", rpcservice.BROADCAST_SERVICE)
	s := new(Service)
	return rpc.RegisterName(rpcservice.BROADCAST_SERVICE, s)
}

func (b *Service) Join(request rpcservice.JoinLeaveReq, reply *rpcservice.RpcResp) error  {
	log.Infof("rpc receive join %+v", request)
	hub.DoRegisterRemoteClient(request.PeerId, request.Addr)
	reply.Success = true
	return nil
}

func (b *Service) Leave(request rpcservice.JoinLeaveReq, reply *rpcservice.RpcResp) error  {
	log.Infof("rpc receive leave %v", request)
	hub.DoUnregister(request.PeerId)
	reply.Success = true
	return nil
}

func (h *Service) Pong(request rpcservice.Ping, reply *rpcservice.Pong) error {
	return nil
}
