package main

import (
	"log"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var EventsCollector *eventsCollector
var hub *Hub

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	log.Printf("UpgradeHTTP")
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}
	//if err := EventsCollector.Add(conn); err != nil {
	//	log.Printf("Failed to add connection %v", err)
	//	conn.Close()
	//}

	//id := r.Form.Get("id")
	//if id == "" {
	//	conn.Close()
	//}

	//msg, op, err := wsutil.ReadClientData(conn)
	//if err != nil {
	//	log.Printf("err: %s", err)
	//}
	//log.Printf("msg: %s %d", string(msg), op)


	defer func() {
		log.Printf("node leave")
		conn.Close()
	}()

	for {
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			// handle error
		}
		err = wsutil.WriteServerMessage(conn, op, msg)
		if err != nil {
			// handle error
		}
	}
	log.Printf("node leave")
}

func main() {
	// Increase resources limitations
	//var rLimit syscall.Rlimit
	//if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
	//	panic(err)
	//}
	//rLimit.Cur = rLimit.Max
	//if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
	//	panic(err)
	//}

	// Enable pprof hooks
	//go func() {
	//	if err := http.ListenAndServe("localhost:6060", nil); err != nil {
	//		log.Fatalf("pprof failed: %v", err)
	//	}
	//}()

	// Start event
	//var err error
	//EventsCollector, err = MkEventsCollector()
	//if err != nil {
	//	panic(err)
	//}

	//go Start()        // TODO 打开

	http.HandleFunc("/", wsHandler)
	if err := http.ListenAndServe("0.0.0.0:80", nil); err != nil {
		log.Fatal(err)
	}

}

func Start() {
	for {
		connections, err := EventsCollector.Wait()
		if err != nil {
			log.Printf("Failed to epoll wait %v", err)
			continue
		}
		log.Printf("6666666666")
		for _, conn := range connections {
			if conn == nil {
				break
			}
			if _, _, err := wsutil.ReadClientData(conn); err != nil {
				if err := EventsCollector.Remove(conn); err != nil {
					log.Printf("Failed to remove %v", err)
				}
				conn.Close()
			} else {
				// This is commented out since in demo usage, stdout is showing messages sent from > 1M connections at very high rate
				//log.Printf("msg: %s", string(conn.LocalAddr()))
				//msg, op, err := wsutil.ReadClientData(conn)
				//if err != nil {
				//	log.Printf("err: %s", err)
				//}
				//log.Printf("msg: %s %d", string(msg), op)
			}
		}
	}
}