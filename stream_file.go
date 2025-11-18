package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	FileStream            = "STREAM_FILE"
	BaseFileStreamSubject = "file.stream"
)

type fileMeta struct {
	Size int64
	Name string
}

func sendFileStream(ctx context.Context, app *NatsConnector, filePath, subject string) error {
	js := app.GetJetStream()
	uid := uuid.New().String()

	fmt.Printf("opening file : '%s'\n", filePath)
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, _ := f.Stat()
	fileSize := stat.Size()
	var fs int64
	buf := bufio.NewReader(f)
	chunk := make([]byte, 1024)

	var fm = fileMeta{
		Size: fileSize,
		Name: f.Name(),
	}
	data, _ := json.Marshal(fm)
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set("action", "new")
	msg.Header.Set("id", uid)
	msg.Header.Set("from", fmt.Sprintf("%s-%s", app.cnf.DC, app.cnf.Name))

	_, err = js.PublishMsg(ctx, msg)
	if err != nil {
		return err
	}

	fmt.Printf("send new file name=%s size=%d id=%s", fm.Name, fm.Size, uid)

	for {
		n, err := buf.Read(chunk)
		if err != nil {
			return err
		}

		msg.Data = chunk[:n]
		msg.Header.Set("action", "chunk")

		_, err = js.PublishMsg(ctx, msg)
		if err != nil {
			return err
		}

		fs += int64(n)
		if fs >= fileSize {
			fmt.Printf("send file finish : %d/%d\n", fs, fileSize)
			break
		}

	}

	return nil
}

func startFileReceiver(ctx context.Context, app *NatsConnector, destPath, receiveSubject string) error {
	js := app.GetJetStream()

	cons, err := js.CreateOrUpdateConsumer(ctx, "FILE", jetstream.ConsumerConfig{
		// AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject:     receiveSubject,
		InactiveThreshold: 1000 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("error in get/create FILE stream consumer : %v", err)
	}

	type fileState struct {
		meta         fileMeta
		writer       *os.File
		receivedSize int64
		totalSize    int64
	}

	var files = make(map[string]*fileState)
	var filesMu sync.Mutex

	cons.Consume(func(msg jetstream.Msg) {
		fmt.Println("debug test")
		h := msg.Headers()
		id := h.Get("id")
		from := h.Get("from")
		action := h.Get("action")

		if id == "" {
			fmt.Println("message without id, ignored")
			return
		}

		switch action {

		// -----------------------
		//  NEW FILE
		// -----------------------
		case "new":
			var fm fileMeta
			if err := json.Unmarshal(msg.Data(), &fm); err != nil {
				fmt.Println("meta decode error:", err)
				return
			}

			filesMu.Lock()
			defer filesMu.Unlock()

			if _, exists := files[id]; !exists {

				f, err := os.Create(path.Join(destPath, fm.Name))
				if err != nil {
					fmt.Println("file create error:", err)
					return
				}

				files[id] = &fileState{
					meta:         fm,
					writer:       f,
					receivedSize: 0,
					totalSize:    fm.Size,
				}

				fmt.Printf("NEW file id=%s name=%s size=%d\n",
					id, fm.Name, fm.Size)
			}

		// -----------------------
		//  FILE CHUNK
		// -----------------------
		case "chunk":

			filesMu.Lock()
			fs, ok := files[id]
			filesMu.Unlock()

			if !ok {
				fmt.Println("chunk received for unknown file id:", id, from)
				time.Sleep(time.Second * 5)
				return
			}

			_, err := fs.writer.Write(msg.Data())
			if err != nil {
				fmt.Println("write error:", err)
				return
			}

			fs.receivedSize += int64(len(msg.Data()))

			fmt.Printf("received %d/%d for id=%s\n",
				fs.receivedSize, fs.totalSize, id)

			if fs.receivedSize == fs.totalSize {
				fmt.Printf("file id=%s completed (%s)\n", id, fs.meta.Name)
				fs.writer.Close()

				filesMu.Lock()
				delete(files, id)
				filesMu.Unlock()
			}

		default:
			fmt.Println("unknown action:", action)
		}
		msg.Ack()
	})

	return nil
}
