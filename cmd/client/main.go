package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	tinykafka "github.com/Stolkerve/tiny-kafka"
	"github.com/Stolkerve/tiny-kafka/sdk"
)

func main() {
	interactive := flag.Bool("it", false, "use the interactive mode for sending events")
	listingOnly := flag.Bool("l", false, "listing only incoming events")
	flag.Parse()

	var cli *sdk.TinyKafkaClient
	if *interactive {
		cli = sdk.NewTinyKafkaClient("localhost", uint16(5678), func(ec tinykafka.EventCommand) {})

		err := cli.CreateConecction()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(-1)
		}

		scanner := bufio.NewScanner(os.Stdin)
		for {
			var eventName, eventJson string
			fmt.Print("Event name: ")
			for scanner.Scan() {
				break
			}
			eventName = scanner.Text()

			fmt.Print("Event json: ")
			for scanner.Scan() {
				break
			}

			eventJson = scanner.Text()

			if !json.Valid([]byte(eventJson)) {
				fmt.Fprintln(os.Stderr, "invalid json")
				continue
			}

			cli.WriteEvent(context.TODO(), tinykafka.EventCommand{
				Name: eventName,
				Data: []byte(eventJson),
			})
		}
	} else if *listingOnly {
		cli = sdk.NewTinyKafkaClient("localhost", uint16(5678), func(ec tinykafka.EventCommand) {
			outEevent := OutgoingEnvent{
				Name: ec.Name,
			}
			if err := json.Unmarshal(ec.Data, &outEevent.Data); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			outJson, err := json.MarshalIndent(outEevent, "", "  ")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			fmt.Println(string(outJson))
		})
		if err := cli.CreateConecction(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(-1)
		}

		for {
		}
	} else {
		var eventInput []byte
		if len(os.Args) > 1 {
			eventInput = []byte(os.Args[1])
		} else {
			fi, err := os.Stdin.Stat()
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(-1)
			}
			if fi.Mode()&os.ModeNamedPipe == 0 {
				// errror
				return
			}

			stdoutBuff := bytes.NewBuffer([]byte{})
			_, err = io.Copy(stdoutBuff, os.Stdin)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(-1)
			}
			eventInput = stdoutBuff.Bytes()
		}

		var outEevent OutgoingEnvent
		if err := json.Unmarshal(eventInput, &outEevent); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(-1)
		}

		cli = sdk.NewTinyKafkaClient("localhost", uint16(5678), func(ec tinykafka.EventCommand) {})
		if err := cli.CreateConecction(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(-1)
		}

		dataBuff, _ := json.Marshal(outEevent.Data)

		cli.WriteEvent(context.TODO(), tinykafka.EventCommand{
			Name: outEevent.Name,
			Data: dataBuff,
		})
	}
}

type OutgoingEnvent struct {
	Name string      `json:"eventName"`
	Data interface{} `json:"eventData"`
}
