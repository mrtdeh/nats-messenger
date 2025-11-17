package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Client struct {
	nc     *nats.Conn
	js     jetstream.JetStream
	kvs    *KeyValue
	closeC chan struct{}
}

var (
	defaultBucketConfigs = []jetstream.KeyValueConfig{
		{
			Bucket:  BUCKET_PERSISTENSE,
			TTL:     0,
			Storage: jetstream.FileStorage,
		}, {
			Bucket:  BUCKET_NODES_HEALTH,
			TTL:     time.Second * 10,
			Storage: jetstream.MemoryStorage,
		},
	}
)

func newClient(natsURL string) (*Client, error) {
	ctx := context.Background()
	c := make(chan struct{})
	// var reconnect_try = 0
	nc, err := nats.Connect(natsURL,
		nats.UserInfo("app", "app"),
		nats.ConnectHandler(func(c *nats.Conn) {
			log.Println("connected to nats : ", c.ConnectedServerName())
		}),
		nats.DisconnectHandler(func(cc *nats.Conn) {
			log.Println("disconencting from nats...", cc.ConnectedServerName())
		}),
		nats.DisconnectErrHandler(func(cc *nats.Conn, err error) {
			log.Println("disconencted error : ", err)
		}),
		nats.ReconnectHandler(func(c *nats.Conn) {
			log.Println("reconnecting to nats : ", c.ConnectedServerName())
		}),
		nats.ReconnectErrHandler(func(cc *nats.Conn, err error) {
			log.Println("reconnecting error : ", cc.ConnectedServerName(), err)
			// reconnect_try += 1
			// if reconnect_try > 5 {
			// 	log.Fatal("all servers is down")
			// }
		}),
	)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}
	if err := waitForLoading(ctx, js, time.Second*20); err != nil {
		return nil, err
	}

	kvs := NewKeyValueStore(js)
	err = kvs.CreateBuckets(ctx, defaultBucketConfigs...)
	if err != nil {
		return nil, fmt.Errorf("CreateBuckets : %s", err)
	}

	if err := ensureChatStream(ctx, js); err != nil {
		return nil, fmt.Errorf("ensureChatStream : %s", err)
	}

	return &Client{nc, js, kvs, c}, nil
}

func waitForLoading(ctx context.Context, js jetstream.JetStream, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		debugPrint("try again to test jetstream")
		ai, err := js.AccountInfo(ctx)
		if err != nil {
			debugPrint(err.Error())
			time.Sleep(time.Second)
			continue
		}
		debugPrint("client connected : ", ai.Domain)
		return nil
	}

	return fmt.Errorf("timout")
}

func ensureChatStream(ctx context.Context, js jetstream.JetStream) error {
	_, err := js.Stream(ctx, "CHAT")
	if err == nil {
		return nil
	}

	_, err = js.CreateStream(ctx, jetstream.StreamConfig{
		Name:      "CHAT",
		Subjects:  []string{"chat.>"},
		Storage:   jetstream.FileStorage,
		Retention: jetstream.LimitsPolicy,
	})
	return err
}

func getNodePath(cnf Config) string {
	return fmt.Sprintf("nodes/%s/%s", cnf.DC, cnf.Name)
}

func Connect(cnf Config) (*NatsConnector, error) {
	c, err := newClient(cnf.NatsURL)
	if err != nil {
		return nil, err
	}

	_, err = c.kvs.Put(
		BUCKET_PERSISTENSE,
		getNodePath(cnf),
		time.Now().String(),
	)
	if err != nil {
		return nil, fmt.Errorf("error in put : %v", err)
	}

	return &NatsConnector{
		cli: c,
		cnf: cnf,
	}, nil
}

func (a *NatsConnector) GetNodes() ([]string, error) {
	nodes, err := a.cli.kvs.Keys(BUCKET_PERSISTENSE, "nodes/")
	if err != nil {
		return nil, err
	}

	heals, err := a.cli.kvs.MapKeys(BUCKET_NODES_HEALTH, "")
	if err != nil {
		return nil, err
	}

	var nodeHeals []string
	for _, n := range nodes {
		nn := strings.Join(strings.Split(n, "/")[1:], "/")
		heal := "nok"
		if ok := heals[nn]; ok {
			heal = "ok"
		}
		nodeHeals = append(nodeHeals, fmt.Sprintf("%s : %s", n, heal))
	}

	return nodeHeals, nil
}

func (a *NatsConnector) GetDCs() ([]string, error) {
	nodes, err := a.GetNodes()
	if err != nil {
		return nil, err
	}
	var dcsm = map[string]bool{}
	var dcs = []string{}
	for _, n := range nodes {
		if n == "" {
			continue
		}
		parts := strings.Split(n, "/")
		if len(parts) < 2 {
			continue
		}
		dc := parts[1]
		if e := dcsm[dc]; !e {
			dcsm[dc] = true
			dcs = append(dcs, dc)
		}
	}

	return dcs, nil
}

func (a *NatsConnector) StartHealthCheck() error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	nodePath := fmt.Sprintf("%s/%s", a.cnf.DC, a.cnf.Name)

	for {
		select {
		case <-ticker.C:
			// Start a heartbeat attempt
			if err := a.sendHeartbeat(nodePath); err != nil {
				// فقط لاگ یا هندل — ولی return نه
				fmt.Println("health check failed:", err)
			}

		case <-a.cli.closeC:
			return nil
		}
	}
}

func (a *NatsConnector) sendHeartbeat(nodePath string) error {
	const (
		maxTimeout = 5 * time.Second        // timeout کلی برای هر تلاش
		baseDelay  = 200 * time.Millisecond // شروع backoff
		maxDelay   = 2 * time.Second        // سقف backoff
	)

	deadline := time.Now().Add(maxTimeout)
	delay := baseDelay

	for {
		// تلاش برای Put
		_, err := a.cli.kvs.Put(BUCKET_NODES_HEALTH, nodePath, []byte(time.Now().Format(time.RFC3339)))
		if err == nil {
			return nil // موفق
		}

		// اگر timeout تمام شده → شکست
		if time.Now().Add(delay).After(deadline) {
			return fmt.Errorf("heartbeat timeout: %w", err)
		}

		// صبر backoff
		time.Sleep(delay)

		// افزایش backoff
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}
}

func (a *NatsConnector) Close() {
	a.cli.nc.Close()
	close(a.cli.closeC)
}

func (a *NatsConnector) ToLocalSubject(subjects ...string) string {
	return strings.ToLower(fmt.Sprintf("%s.%s.%s", a.cnf.DC, a.cnf.Name, strings.Join(subjects, ".")))
}

func (a *NatsConnector) ToLocalStream(streamNames ...string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s_%s", a.cnf.DC, a.cnf.Name, strings.Join(streamNames, "_")))
}

func (a *NatsConnector) GetJetStream() jetstream.JetStream {
	return a.cli.js
}
