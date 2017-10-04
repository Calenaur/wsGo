package server

import (
	"fmt"
	"net"

	"github.com/Calenaur/wsGo/client"
)

type Server struct {
	IdIndex          int
	MaxClients       int
	MaxMessageSize   int
	MaxHandshakeTime int
	LastUID          int
	Clients          []*client.Client
	Ip               string
	Port             int
	Connected        bool
	EventListeners   []client.EventListener
}

func New(maxClients int, maxMessageSize int, maxHandshakeTime int) *Server {
	return &Server{0, maxClients, maxMessageSize, maxHandshakeTime, 0, []*client.Client{}, "", 0, false, nil}
}

func (s *Server) AddClient(client *client.Client) {
	if len(s.Clients) < s.MaxClients {
		s.Clients = append(s.Clients, client)
	}
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

func (s *Server) Start(ip string, port int) {
	sock, err := net.Listen("tcp", fmt.Sprint(ip, ":", port))
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
		}
		go s.HandleConnection(&conn)
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
