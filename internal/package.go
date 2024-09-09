package internal

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	tinykafka "github.com/Stolkerve/tiny-kafka"
)

type Package struct {
	Type tinykafka.Command
	Body interface{}
}

type BroadcastPackage struct {
	Pkg   Package
	Addrs string
}

func EncodePackage(p Package) ([]byte, error) {
	jsonBuff, err := json.Marshal(p.Body)
	if err != nil {
		return nil, err
	}

	buff := make([]byte, 1+4)
	binary.LittleEndian.PutUint32(buff, uint32(len(jsonBuff)))
	buff[4] = byte(p.Type)

	buff = append(buff, jsonBuff...)

	return buff, nil
}

func DecodePackage(packageBuff []byte) (*Package, error) {
	commandType := tinykafka.Command(packageBuff[0])
	pkg := &Package{
		Type: commandType,
	}

	switch commandType {
	case tinykafka.EventCommandType:
		var eventCommand tinykafka.EventCommand
		if err := json.Unmarshal(packageBuff[1:], &eventCommand); err != nil {
			return nil, err
		}
		pkg.Body = eventCommand
	default:
		panic(fmt.Sprintf("unexpected pgk.Command: %#v", commandType))
	}

	return pkg, nil
}
