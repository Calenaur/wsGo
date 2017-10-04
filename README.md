# wsGo
A websocket server implentation written in Golang

```go 
package main

import (
	"fmt"

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
	ml := client.EventListener(&mainListener{0})
	server := server.New(32, 512, 60)
	server.AddListener(ml)
	server.Start("", 2302)
}
```