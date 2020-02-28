package hub

import "sync"

type Hub struct {

	clients sync.Map

	ClientNum int64            //count of client

}
