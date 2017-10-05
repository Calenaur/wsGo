package server

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/Calenaur/wsGo/client"
)

type Server struct {
	IdIndex        int
	LastUID        int
	Clients        []*client.Client
	Ip             string
	Port           int
	Connected      bool
	EventListeners []client.EventListener
}

func New() *Server {
	return &Server{0, 0, []*client.Client{}, "", 0, false, nil}
}

func (s *Server) AddClient(client *client.Client) {
	s.Clients = append(s.Clients, client)
}

func (s *Server) AddListener(el client.EventListener) {
	s.EventListeners = append(s.EventListeners, el)
}

func (s *Server) Cleanup() {
	for i := 0; i < len(s.Clients); i++ {
		if !s.Clients[i].IsConnected {
			s.RemoveClient(i)
			i--
		}
	}
}

func (s *Server) Start(ip string, port int, config *tls.Config) {
	var sock net.Listener
	var err error
	if config != nil {
		sock, err = tls.Listen("tcp", fmt.Sprint(ip, ":", port), config)
	} else {
		sock, err = net.Listen("tcp", fmt.Sprint(ip, ":", port))
	}
	if err != nil {
		s.Connected = false
		return
	}
	s.Ip = ip
	s.Port = port
	s.Connected = true
	for {
		conn, err := sock.Accept()
		if err != nil {
			fmt.Println(err)
		} else {
			go s.HandleConnection(&conn)
		}
	}
	s.Connected = false
}

func (s *Server) HandleConnection(conn *net.Conn) {
	s.Cleanup()
	c := client.New(s.IdIndex, conn, s.EventListeners)
	if c != nil {
		c.RaiseOnConnect()
		s.IdIndex++
		s.AddClient(c)
	}
}

func (s *Server) RemoveClient(i int) {
	s.Clients[i].RaiseDisconnect()
	s.Clients = append(s.Clients[:i], s.Clients[i+1:]...)
}

func (s *Server) RemoveClientByRefrence(c *client.Client) {
	for i, cl := range s.Clients {
		if cl == c {
			s.RemoveClient(i)
			return
		}
	}
}
