package main

import (
	"bufio"
	"os"
	"strings"
)

func StartRealtimeInput() <-chan string {
	inputChan := make(chan string)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				close(inputChan)
				return
			}
			text = strings.TrimSpace(text)
			inputChan <- text
		}
	}()

	return inputChan
}

// func testOrderGuarantee1(js nats.JetStreamContext) {
// 	streamName := "TEST"
// 	subjectName := "test.orders2"

// 	_, err := js.StreamInfo(streamName)
// 	if err != nil {
// 		_, err = js.AddStream(&nats.StreamConfig{
// 			Name:     streamName,
// 			Subjects: []string{subjectName},
// 			Storage:  nats.FileStorage,
// 			Replicas: 1,
// 		})
// 		if err != nil {
// 			panic(err)
// 		}
// 	}

// 	go func() {
// 		sub, err := js.PullSubscribe(subjectName, "demo-consumer")

// 		if err != nil {
// 			panic(err)
// 		}

// 		for {
// 			msgs, err := sub.Fetch(10, nats.MaxWait(2*time.Second))
// 			if err != nil {
// 				continue
// 			}
// 			for _, msg := range msgs {
// 				fmt.Printf("📩 Received: %s\n", string(msg.Data))
// 				msg.Ack()
// 			}
// 		}
// 	}()

// 	for i := 1; i <= 100; i++ {

// 		p := payload{i, "data"}
// 		msgBytes, _ := json.Marshal(p)
// 		// ctx := context.Background()
// 		ack, err := js.Publish(subjectName, msgBytes)
// 		if err != nil {
// 			log.Fatalf("publish error %v", err)
// 		}

// 		fmt.Printf("Published: %d , ackSeq: %v\n", i, ack.Sequence)
// 		// time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
// 		time.Sleep(20 * time.Millisecond)
// 	}
// }

// func testOrderGuarantee2(js nats.JetStreamContext) {
// 	streamName := "TEST2"
// 	subjectName := "test.orders"

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

// 	var l = &sync.Mutex{}
// 	var wg sync.WaitGroup

// 	type payload struct {
// 		Num int
// 		Msg string
// 	}
// 	var sequense = map[string]struct{}{}

// 	for i := 1; i <= 100; i++ {
// 		wg.Add(1)
// 		go func(seq int) {
// 			defer wg.Done()
// 			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
// 			p := payload{seq, "data"}
// 			msgBytes, _ := json.Marshal(p)

// 			// _, err := js.PublishAsync(subjectName, msgBytes)
// 			ack, err := js.Publish(subjectName, msgBytes)
// 			if err != nil {
// 				log.Fatalf("publish error %v", err)
// 			}

// 			fmt.Println("published : ", fmt.Sprintf("msg: %s, num: %d", p.Msg, p.Num))
// 			l.Lock()
// 			key := fmt.Sprintf("%d:%d", ack.Sequence, p.Num)
// 			sequense[key] = struct{}{}
// 			l.Unlock()
// 		}(i)
// 	}

// 	// <-js.PublishAsyncComplete()
// 	wg.Wait()

// 	sub, err := js.PullSubscribe(subjectName, "tester", nats.DeliverAll())
// 	if err != nil {
// 		log.Fatalf("PullSubscribe error %v", err)
// 	}

// 	msgs, err := sub.Fetch(100)
// 	if err != nil {
// 		log.Fatalf("Fetch error %v", err)
// 	}

// 	fmt.Println("len(sequense) : ", len(sequense))
// 	fmt.Println("len(msgs) : ", len(msgs))

// 	failed := false
// 	for _, msg := range msgs {
// 		var p payload
// 		json.Unmarshal(msg.Data, &p)
// 		meta, _ := msg.Metadata()
// 		key := fmt.Sprintf("%d:%d", meta.Sequence.Stream, p.Num)
// 		if _, ok := sequense[key]; !ok {
// 			failed = true
// 			fmt.Printf("ORDER VIOLATION: key not found => %s\n", key)
// 		}

// 	}
// 	if !failed {
// 		fmt.Println("All messages in correct sequence order ✅")
// 	}
// }
