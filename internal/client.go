package internal

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"

	tinykafka "github.com/Stolkerve/tiny-kafka"
)

type Client struct {
	Socket             net.Conn
	Addrs              string
	CloseConectionChan chan string
	BroadcastEventChan chan BroadcastPackage
}

func NewClient(socket net.Conn, addrs string, broadcastEventChan chan BroadcastPackage, closeConectionChan chan string) *Client {
	return &Client{
		Socket:             socket,
		Addrs:              addrs,
		CloseConectionChan: closeConectionChan,
		BroadcastEventChan: broadcastEventChan,
	}
}

func (c *Client) HandleConection() {
	bodySizeBuffer := make([]byte, 4)
	for {
		bodySizeBuffer[0] = 0
		bodySizeBuffer[1] = 0
		bodySizeBuffer[2] = 0
		bodySizeBuffer[3] = 0

		if _, err := c.Socket.Read(bodySizeBuffer); err != nil {
			c.CloseConectionChan <- c.Addrs
			return
		}
		packageSize := binary.LittleEndian.Uint32(bodySizeBuffer)
		packageBuff := make([]byte, packageSize+1)
		if _, err := c.Socket.Read(packageBuff); err != nil {
			c.CloseConectionChan <- c.Addrs
			return
		}

		pkg, err := DecodePackage(packageBuff)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			c.CloseConectionChan <- c.Addrs
			return
		}

		switch pkg.Type {
		case tinykafka.EventCommandType:
			c.BroadcastEventChan <- BroadcastPackage{
				Addrs: c.Addrs,
				Pkg:   *pkg,
			}
		}
	}
}
