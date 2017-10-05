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
	fmt.Println("Client joined count:", ml.clients)
}

func (ml *mainListener) OnMessage(c *client.Client, f *frame.Frame) {
	fmt.Println("Client message:", string(f.Data))
}

func (ml *mainListener) OnDisconnect(c *client.Client) {
	ml.clients--
	fmt.Println("Client left count:", ml.clients)
}

func main() {
	server := server.New()
	server.AddListener(&mainListener{0})
	server.Start("", 2302, nil)
}
```