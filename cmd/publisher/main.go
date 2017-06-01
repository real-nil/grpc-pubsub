package main

import (
	"context"
	"flag"
	"github.com/real-nil/grpc-pubsub/proto/pubsub"
	"google.golang.org/grpc"
	"log"
	"time"
)

var (
	pubsubAddr = flag.String(`pubsub`, ``, `set pubsub address`)
	topic      = flag.String(`topic`, ``, `set topic for publications`)
	msg        = flag.String(`msg`, ``, `set msg for publish`)
	count      = flag.Int(`count`, 0, `set count for send message if 0 = infinity`)
	timeout    = flag.Duration(`timeout`, 10*time.Second, `set how log do it, other side of count`)
	parallel   = flag.Int(`parallel`, 16, `set parallel process`)
)

var Q chan func()

func main() {
	flag.Parse()

	var (
		cancel context.CancelFunc
		ctx    context.Context
	)

	ctx, cancel = context.WithCancel(context.Background())
	if *timeout > 0 {
		print(`with timeout`)
		ctx, cancel = context.WithTimeout(ctx, *timeout)
	}

	defer cancel()

	conn, err := grpc.DialContext(ctx, *pubsubAddr, grpc.WithInsecure(), grpc.WithBackoffMaxDelay(*timeout), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	Q = make(chan func(), *count)
	runWorkers(ctx, *parallel)

	pubsubCli := pubsub.NewPubSubClient(conn)
	log.Printf(`start pbclish to %s/%s(%s) with count %d or timeout %s \n`, *pubsubAddr, *topic, *msg, *count, *timeout)
	log.Printf(`parallel =%d`, *parallel)
	in := &pubsub.Data{
		Topic:   *topic,
		Payload: []byte(*msg),
	}
	job := func() {
		_, err := pubsubCli.Publish(ctx, in)
		if err != nil {
			log.Printf(`can't publish: %s`, err)
		}
	}
	for {
		Q <- job
	}
}

func runWorkers(ctx context.Context, size int) {
	for i := 0; i < size; i++ {
		go func() {
			select {
			case job := <-Q:
				job()
			case <-ctx.Done():
				return
			}
		}()
	}
}
