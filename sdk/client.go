package sdk

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"

	"github.com/Stolkerve/tiny-kafka"
	"github.com/Stolkerve/tiny-kafka/internal"
)

type EventsHandlerFunc func(tinykafka.EventCommand)

type TinyKafkaClient struct {
	ServerAddress string
	ServerPort    uint16
	socket        net.Conn
	eventsHandler EventsHandlerFunc
}

func NewTinyKafkaClient(addrs string, port uint16, eventsHandler EventsHandlerFunc) *TinyKafkaClient {
	return &TinyKafkaClient{
		ServerAddress: addrs,
		ServerPort:    port,
		eventsHandler: eventsHandler,
	}
}

func (c *TinyKafkaClient) CreateConecction() error {
	socket, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.ServerAddress, c.ServerPort))
	if err != nil {
		return err
	}

	c.socket = socket

	go c.listingIncomingEvents()

	return nil
}

func (c *TinyKafkaClient) WriteEvent(ctx context.Context, event tinykafka.EventCommand) error {
	_, ok := ctx.Deadline()
	if !ok {
		// context.WithTimeout(ctx, time.Second*30)
	}

	pkgBuff, err := internal.EncodePackage(internal.Package{
		Type: tinykafka.EventCommandType,
		Body: event,
	})
	if err != nil {
		return err
	}

	_, err = c.socket.Write(pkgBuff)
	if err != nil {
		return err
	}

	return nil
}

func (c *TinyKafkaClient) Close() error {
	return nil
}

func (c *TinyKafkaClient) listingIncomingEvents() {
	bodySizeBuffer := make([]byte, 4)
	for {
		bodySizeBuffer[0] = 0
		bodySizeBuffer[1] = 0
		bodySizeBuffer[2] = 0
		bodySizeBuffer[3] = 0
		_, err := c.socket.Read(bodySizeBuffer)
		if err != nil {
			continue
		}

		packageSize := binary.LittleEndian.Uint32(bodySizeBuffer)
		packageBuff := make([]byte, packageSize+1)
		if _, err := c.socket.Read(packageBuff); err != nil {
			continue
		}

		pkg, err := internal.DecodePackage(packageBuff)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		switch pkg.Type {
		case tinykafka.EventCommandType:
			event := pkg.Body.(tinykafka.EventCommand)
			c.eventsHandler(event)
		}
	}
}
