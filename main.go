package main

import (
	"fmt"
	"sync"

	"github.com/Calenaur/wsGo/client"
	"github.com/Calenaur/wsGo/frame"
	"github.com/Calenaur/wsGo/server"
)

type mainListener struct {
	clients int
}

func (ml *mainListener) OnConnect(c *client.Client) {
	ml.clients++
	fmt.Println("Player joined count:", ml.clients)
}

func (ml *mainListener) OnMessage(c *client.Client, f *frame.Frame) {
	fmt.Println("Player message:", string(f.Data))
	r := frame.NewFromBinary(frame.OPCODE_PING, []byte{})
	c.Send(r)
}

func (ml *mainListener) OnDisconnect(c *client.Client) {
	ml.clients--
	fmt.Println("Player left count:", ml.clients)
}

func main() {
	var wg sync.WaitGroup
	var el client.EventListener
	ml := mainListener{0}
	el = &ml
	server := server.New(32, 512, 60)
	server.AddListener(el)
	wg.Add(1)
	go server.Connect("", 2302, &wg)
	wg.Wait()
}
