package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	BUCKET_PERSISTENSE   = "BUCKET_PERSISTENSE"
	BUCKET_NODES_HISTORY = "NODES_HISTORY"
	BUCKET_NODES_HEALTH  = "NODES_HEALTH"
)

type NatsConnector struct {
	cli *Client
	cnf Config
}

func main() {
	cnf := LoadConfiguration()
	ctx := context.Background()
	sh := NewShell()

	nc, err := Connect(cnf)
	if err != nil {
		log.Fatal("error in connect : ", err)
	}
	defer nc.Close()

	receiveFileSubject := fmt.Sprintf("file.%s.%s", cnf.DC, cnf.Name)
	go func() {
		err := startFileReceiver(ctx, nc, "/tmp", receiveFileSubject)
		if err != nil {
			log.Fatal("startFileReceiver error :", err)
		}
	}()

	go func() {
		if err := nc.StartHealthCheck(); err != nil {
			log.Fatal("health-check error : ", err)
		}
		log.Println("health-check closed")
	}()

	receiveSubject := fmt.Sprintf("chat.%s.%s", cnf.DC, cnf.Name)
	cons, err := nc.cli.js.CreateOrUpdateConsumer(ctx, "CHAT", jetstream.ConsumerConfig{
		InactiveThreshold: 10 * time.Minute,
		FilterSubject:     receiveSubject,
	})
	if err != nil {
		log.Fatal("failed to create consumer : ", err)
	}

	_, err = cons.Consume(func(msg jetstream.Msg) {

		fmt.Println(string(msg.Data()))
		msg.Ack()

	})
	if err != nil {
		log.Fatal("failed to consume : ", err)
	}

	fmt.Println("Listen for path : ", receiveSubject)

	sh.Run(func(sh *Shell, command, args string) error {
		p := strings.Split(sh.path, "/")
		switch command {
		case "goto":
			_, err := nc.cli.kvs.Get(BUCKET_PERSISTENSE, args)
			if err != nil {
				return err
			}

			sh.Goto(args)
		case "send":
			if sh.path == "root" || sh.path == "" {
				return fmt.Errorf("not any path selected")
			}

			subject := "chat." + strings.Join(p[1:], ".")
			_, err := nc.cli.js.Publish(ctx, subject, []byte(args))
			if err != nil {
				return err
			}
		case "sendfile":
			if sh.path == "root" || sh.path == "" {
				return fmt.Errorf("not any path selected")
			}

			subject := "file." + strings.Join(p[1:], ".")
			err := sendFileStream(ctx, nc, args, subject)
			if err != nil {
				return err
			}

		case "nodes":
			nodes, err := nc.GetNodes()
			if err != nil {
				return fmt.Errorf("error in get  nodes : %v", err)
			}
			fmt.Println(strings.Join(nodes, "\n"))
		case "dcs":
			dcs, err := nc.GetDCs()
			if err != nil {
				return fmt.Errorf("error in get dcs : %v", err)
			}
			fmt.Println(strings.Join(dcs, "\n"))

		case "exit", "quit":
			fmt.Println("Bye!")
			os.Exit(0)
		default:
			fmt.Println("Unknown command:", command)
		}
		return nil
	})

}

func debugPrint(strs ...string) {
	now := time.Now().Format("2006-01-02 15:04:05.000") // yyyy-mm-dd hh:mm:ss.mmm
	fmt.Printf("[%s] %s\n", now, strings.Join(strs, " "))
}
