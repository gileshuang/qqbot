package main

import (
	"qqbot/lib/qblog"
	"qqbot/lib/ws"
)

func main() {
	qblog.Log.SetDebug(true)
	qblog.Log.Info("starting websocket server...")
	wss := ws.NewWsServer("0.0.0.0:6701")
	wss.Start()
}
