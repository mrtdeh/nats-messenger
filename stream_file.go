package main

// const (
// 	FileStream            = "STREAM_FILE"
// 	BaseFileStreamSubject = "file.stream"
// )

// func sendFileStream(app *NatsConnector, filePath string) error {
// 	streamName := app.ToLocalStream(FileStream)
// 	subjectName := app.ToLocalSubject(BaseFileStreamSubject)
// 	js := app.GetJetStream()

// 	_, err := js.StreamInfo(streamName)
// 	if err != nil {
// 		_, err := js.AddStream(&nats.StreamConfig{
// 			Name:     streamName,
// 			Subjects: []string{subjectName},
// 			Storage:  nats.FileStorage,
// 			Replicas: 1,
// 		})
// 		if err != nil {
// 			log.Fatalf("add stream error %v", err)
// 		}
// 	}

// 	f, err := os.Open(filePath)
// 	if err != nil {
// 		return fmt.Errorf("err2 : %v", err)
// 	}
// 	defer f.Close()

// 	// info, err := objStore.Put(&nats.ObjectMeta{Name: objectName}, f)
// 	// if err != nil {
// 	// 	return fmt.Errorf("err3 : %v", err)
// 	// }
// 	// fmt.Println("Uploaded:", info)
// 	return nil
// }

// func getFileStream(nc *nats.Conn, objectName, destPath string) error {
// 	js, err := nc.JetStream()
// 	if err != nil {
// 		return err
// 	}

// 	objStore, err := js.ObjectStore(objectName)
// 	if err != nil {
// 		return fmt.Errorf("err1 : %v", err)
// 	}

// 	out, err := os.Create(destPath)
// 	if err != nil {
// 		return fmt.Errorf("err4 : %v", err)
// 	}
// 	defer out.Close()

// 	obj, err := objStore.Get(objectName)
// 	if err != nil {
// 		return fmt.Errorf("err5 : %v", err)
// 	}

// 	_, err = io.Copy(out, obj)
// 	if err != nil {
// 		return fmt.Errorf("err6 : %v", err)
// 	}
// 	fmt.Println("Downloaded OK")
// 	return nil
// }

// type MetaFile struct {
// 	Hash string
// 	Name string
// 	Size int
// }

// func recvStream(app *NatsConnector, destPath string) error {
// 	nc := app.cli.nc
// 	metaChan := make(chan *nats.Msg, 64)

// 	// Subscribe to metadata messages
// 	metaSub, err := nc.ChanSubscribe(app.ToLocalSubject("file.meta"), metaChan)
// 	if err != nil {
// 		return fmt.Errorf("meta subscribe: %w", err)
// 	}
// 	defer metaSub.Unsubscribe()

// 	fmt.Println("📡 Listening for file meta...")

// 	for msg := range metaChan {
// 		var mt MetaFile
// 		if err := json.Unmarshal(msg.Data, &mt); err != nil {
// 			msg.Respond([]byte("error: bad meta json"))
// 			continue
// 		}

// 		// Confirm we got the meta
// 		msg.Respond([]byte("ok"))
// 		fmt.Printf("🧩 New file request: %s (%d bytes) [hash:%s]\n", mt.Name, mt.Size, mt.Hash)

// 		// Launch a new goroutine for this file
// 		go func(mt MetaFile) {
// 			chunkSubj := app.ToLocalSubject("file.chunk", mt.Hash)
// 			chunkChan := make(chan *nats.Msg, 512)

// 			sub, err := nc.ChanSubscribe(chunkSubj, chunkChan)
// 			if err != nil {
// 				fmt.Println("❌ cannot subscribe to chunks:", err)
// 				return
// 			}
// 			defer sub.Unsubscribe()

// 			dest := filepath.Join(destPath, mt.Name)
// 			out, err := os.Create(dest)
// 			if err != nil {
// 				fmt.Println("❌ cannot create file:", err)
// 				return
// 			}
// 			defer out.Close()

// 			var received int
// 			timeout := time.NewTimer(30 * time.Second)

// 			for {
// 				select {
// 				case chunk := <-chunkChan:
// 					if chunk == nil {
// 						continue
// 					}
// 					n, err := out.Write(chunk.Data)
// 					if err != nil {
// 						fmt.Println("❌ write failed:", err)
// 						return
// 					}
// 					received += n
// 					fmt.Printf("⬇️  [%s] %d/%d bytes\n", mt.Name, received, mt.Size)

// 					if received >= mt.Size {
// 						fmt.Printf("✅ File complete: %s\n", dest)
// 						err := nc.Publish(app.ToLocalSubject("file.done", mt.Hash), []byte(dest))
// 						if err != nil {
// 							fmt.Println("❌ write failed:", err)
// 							return
// 						}
// 						return
// 					}
// 					timeout.Reset(30 * time.Second)

// 				case <-timeout.C:
// 					fmt.Printf("⚠️ Timeout waiting for chunks of %s\n", mt.Name)
// 					return
// 				}
// 			}
// 		}(mt)
// 	}
// 	return nil
// }

// func sendStream(app *NatsConnector, filePath string, progress *float64) (string, error) {
// 	nc := app.cli.nc

// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return "", fmt.Errorf("open file: %w", err)
// 	}
// 	defer file.Close()

// 	info, err := file.Stat()
// 	if err != nil {
// 		return "", fmt.Errorf("stat file: %w", err)
// 	}

// 	hash := sha1.New()
// 	if _, err := io.Copy(hash, file); err != nil {
// 		return "", fmt.Errorf("hash: %w", err)
// 	}
// 	file.Seek(0, 0) // برگرد به اول فایل
// 	hashStr := fmt.Sprintf("%x", hash.Sum(nil))

// 	meta := MetaFile{
// 		Name: info.Name(),
// 		Size: int(info.Size()),
// 		Hash: hashStr,
// 	}

// 	metaBytes, _ := json.Marshal(meta)
// 	metaSubj := app.ToLocalSubject("file.meta")
// 	chunkSubj := app.ToLocalSubject(fmt.Sprintf("file.chunk.%s", hashStr))
// 	doneSubj := app.ToLocalSubject(fmt.Sprintf("file.done.%s", hashStr))

// 	// Send file request
// 	fmt.Printf("📤 Sending meta: %s (%d bytes) [%s]\n", meta.Name, meta.Size, meta.Hash)
// 	msg, err := nc.Request(metaSubj, metaBytes, 5*time.Second)
// 	if err != nil {
// 		return "", fmt.Errorf("meta send failed: %w", err)
// 	}
// 	if string(msg.Data) != "ok" {
// 		return "", fmt.Errorf("receiver not ready: %s", string(msg.Data))
// 	}
// 	fmt.Println("✅ Receiver ready, sending chunks...")

// 	// Send Chunks
// 	const chunkSize = 64 * 1024 // 64KB
// 	buf := make([]byte, chunkSize)
// 	sent := 0

// 	for {
// 		n, err := file.Read(buf)
// 		if n > 0 {
// 			err := nc.Publish(chunkSubj, buf[:n])
// 			if err != nil {
// 				return "", fmt.Errorf("publish chunk: %w", err)
// 			}
// 			sent += n
// 			p := float64(sent) / float64(meta.Size) * 100
// 			if progress != nil {
// 				*progress = p
// 			}

// 			fmt.Printf("📦 Sent chunk: %d/%d bytes\n", sent, meta.Size)
// 		}
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			return "", fmt.Errorf("read file: %w", err)
// 		}
// 	}

// 	// Wait for Done
// 	doneChan := make(chan *nats.Msg, 1)
// 	sub, err := nc.ChanSubscribe(doneSubj, doneChan)
// 	if err != nil {
// 		return "", fmt.Errorf("subscribe done: %w", err)
// 	}
// 	defer sub.Unsubscribe()

// 	var destPath string
// 	select {
// 	case msg := <-doneChan:
// 		if string(msg.Data) != "" {
// 			destPath = string(msg.Data)
// 			fmt.Printf("🎉 File sent successfully: %s\n", meta.Name)
// 		} else {
// 			fmt.Printf("⚠️ Receiver returned unexpected message: %s\n", string(msg.Data))
// 			return "", fmt.Errorf("destination done error: %w", err)
// 		}
// 	case <-time.After(10 * time.Second):
// 		fmt.Println("⏱️ Timeout waiting for done-ok")
// 		return "", fmt.Errorf("destination done timeout: %w", err)
// 	}

// 	return destPath, nil
// }
