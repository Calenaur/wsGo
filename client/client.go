package client

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	"github.com/Calenaur/wsGo/frame"
	"github.com/Calenaur/wsGo/framebuilder"
	"github.com/Calenaur/wsGo/http"
)

const magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type EventListener interface {
	OnConnect(c *Client)
	OnMessage(c *Client, f *frame.Frame)
	OnDisconnect(c *Client)
}

type Client struct {
	conn           *net.Conn
	eventListeners []EventListener
	Id             int
	IsValid        bool
	IsConnected    bool
}

func New(id int, conn *net.Conn, eventListeners []EventListener) *Client {
	c := Client{conn, eventListeners, id, false, true}
	c.Handshake()
	if c.IsValid {
		return &c
	} else {
		return nil
	}
}

func (c *Client) Handshake() {
	var hash string
	h := *c.parseHeaders()
	if k, ok := h["Sec-WebSocket-Key"]; ok {
		s := sha1.New()
		s.Write([]byte(fmt.Sprint(k, magic)))
		result := s.Sum(nil)
		hash = base64.StdEncoding.EncodeToString(result)
	}
	c.acceptKey(hash)
	go c.listen()
}

func (c *Client) listen() {
	conn := *c.conn
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	fb := framebuilder.New()
	for c.IsConnected {
		b, err := rw.ReadByte()
		if err != nil {
			c.RaiseDisconnect()
			break
		}
		fb.Write(b)
		if fb.Available != 0 {
			f := fb.TakeFrame()
			fmt.Println("Frame recieved:", f.Opcode)
			switch f.Opcode {
			case frame.OPCODE_CONT:
			case frame.OPCODE_TEXT:
				c.RaiseOnMessage(f)
			case frame.OPCODE_BINARY:
				c.RaiseOnMessage(f)
			case frame.OPCODE_CLOSE:
				c.RaiseDisconnect()
			case frame.OPCODE_PING:
				r := frame.NewFromBinary(frame.OPCODE_PONG, []byte{})
				rw.Write(r.ToBytes())
				rw.Flush()
			case frame.OPCODE_PONG:
			}
		}
	}
}

func (c *Client) parseHeaders() *map[string]string {
	r := bufio.NewReader(*c.conn)
	h := make(map[string]string)
	var buf []string
	for {
		line, err := r.ReadString('\n')
		if err == nil {
			if line == "\r\n" {
				return &h
			}
			buf = strings.Split(strings.Trim(line, "\r\n"), ": ")
			if len(buf) == 2 {
				h[buf[0]] = buf[1]
			}
		} else {
			break
		}
	}
	return &h
}

func (c *Client) RaiseOnConnect() {
	for _, el := range c.eventListeners {
		el.OnConnect(c)
	}
}

func (c *Client) RaiseOnMessage(f *frame.Frame) {
	for _, el := range c.eventListeners {
		el.OnMessage(c, f)
	}
}

func (c *Client) RaiseDisconnect() {
	if c.IsConnected {
		c.Disconnect()
		for _, el := range c.eventListeners {
			el.OnDisconnect(c)
		}
	}
}

func (c *Client) Send(f *frame.Frame) {
	conn := *c.conn
	w := bufio.NewWriter(conn)
	w.Write(f.ToBytes())
	w.Flush()
}

func (c *Client) acceptKey(k string) {
	w := bufio.NewWriter(*c.conn)
	w.Write(http.Accept(k))
	w.Flush()
	c.IsValid = true
}

func (c *Client) Disconnect() {
	(*c.conn).Close()
	c.IsConnected = false
}
