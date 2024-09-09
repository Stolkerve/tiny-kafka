package internal

import (
	"fmt"
	"log"
	"net"
	"os"
)

type Server struct {
	Address            string
	Clients            map[string]*Client
	NewConectionChan   chan net.Conn
	CloseConectionChan chan string
	BroadcastEventChan chan BroadcastPackage
}

func NewServer(address string) Server {
	return Server{
		Address:            address,
		Clients:            make(map[string]*Client),
		NewConectionChan:   make(chan net.Conn),
		CloseConectionChan: make(chan string),
		BroadcastEventChan: make(chan BroadcastPackage),
	}
}

func (s *Server) ListenConections() error {
	listener, err := net.Listen("tcp", "localhost:5678")
	if err != nil {
		return err
	}

	go s.HandlerStuff()

	for {
		socket, err := listener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}

		s.NewConectionChan <- socket
	}
}

func (s *Server) HandlerStuff() {
	for {
		select {
		case socket := <-s.NewConectionChan:
			addrs := socket.RemoteAddr().String()
			c := NewClient(socket, addrs, s.BroadcastEventChan, s.CloseConectionChan)
			s.Clients[addrs] = c
			go c.HandleConection()
			log.Printf("new client connected at address %s\n", addrs)
		case addrs := <-s.CloseConectionChan:
			s.Clients[addrs].Socket.Close()
			delete(s.Clients, addrs)
			log.Printf("client %s disconected\n", addrs)
		case broadcastPkg := <-s.BroadcastEventChan:
			for _, c := range s.Clients {
				if c.Addrs == broadcastPkg.Addrs {
					continue
				}
				pkg, _ := EncodePackage(broadcastPkg.Pkg)
				if _, err := c.Socket.Write(pkg); err != nil {
					s.CloseConectionChan <- c.Addrs
				}
			}
		}
	}
}
