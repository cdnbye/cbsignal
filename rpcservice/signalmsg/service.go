package signal

import (
	"cbsignal/handler"
	"cbsignal/hub"
	"cbsignal/rpcservice"
	"encoding/json"
	"fmt"
	"github.com/lexkong/log"
	"net/rpc"
)

type Service struct {

}

func RegisterSignalService() error {
	log.Infof("register rpc service %s", rpcservice.SIGNAL_SERVICE)
	s := new(Service)
	return rpc.RegisterName(rpcservice.SIGNAL_SERVICE, s)
}

func (b *Service) Signal(request rpcservice.SignalReq, reply *rpcservice.RpcResp) error  {
	req := handler.SignalResp{}
	if err := json.Unmarshal(request.Data, &req);err != nil {
		return err
	}
	log.Infof("rpc receive signal from %s", req.FromPeerId)
	cli, ok := hub.GetClient(req.FromPeerId)
	if !ok {
		// 节点不存在
		reply.Success = false
		reply.Reason = fmt.Sprintf("peer %s not found", req.FromPeerId)
	} else {
		reply.Success = true
		signalMsg := handler.SignalMsg{
			Action: req.Action,
			ToPeerId: request.ToPeerId,
			Data: req.Data,
		}
		hdr, err := handler.NewHandlerMsg(signalMsg, cli)
		if err != nil {
			return err
		}
		hdr.Handle()
	}
	return nil
}
