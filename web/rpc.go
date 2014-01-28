package web

import (
	"log"
	"net"
	"net/rpc"
	. "github.com/lox/opencoindata/core"
)

var rpcClient *rpc.Client

type TradeServer struct {
	handler func(Trade)
}

func (t *TradeServer) SendTrade(trade *Trade, reply *bool) error {
	t.handler(*trade)
	*reply = true
	return nil
}

func NewTradeServer(handler func(Trade)) (*TradeServer, error) {
	t := &TradeServer{handler}
	rpc.Register(t)
	log.Printf("Listening on tcp://localhost:9999 for trades")
	ln, err := net.Listen("tcp", "localhost:9999")
	if err != nil {
		log.Fatalf("Failed to register rpc server: %v", err)
		return nil, err
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(c)
	}
	return t, nil
}

func SendTrade(t Trade) error {
	if rpcClient == nil {
		c, err := rpc.Dial("tcp", "127.0.0.1:9999")
		if err != nil {
			return err
		}
		rpcClient = c
	}

	var result bool
	err := rpcClient.Call("TradeServer.SendTrade", t, &result)
	if err != nil {
		log.Printf("TradeServer.SendTrade Error: %v", err)
		return err
	}

	return nil
}
