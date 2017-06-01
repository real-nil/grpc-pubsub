package main

import (
	"context"
	"flag"
	"github.com/real-nil/grpc-pubsub/proto/pubsub"
	"google.golang.org/grpc"
	"io"
	"log"
	"sync/atomic"
	"time"
)

var (
	pubsubAddr = flag.String(`pubsub`, `:3456`, `set pubsub address`)
	topic      = flag.String(`topic`, `hellogrpc`, `set topic for subscribe`)
	timeout    = flag.Duration(`timeout`, 1*time.Minute, `set how log do it, other side of count`)
	parallel   = flag.Int(`parallel`, 16, `set parallel subscribers`)
)

var Q chan func()

func main() {
	flag.Parse()

	var (
		cancel  context.CancelFunc
		ctx     context.Context
		counter int64
	)

	defer func() {
		println(`total recv msg count = `, atomic.LoadInt64(&counter))
	}()

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

	Q = make(chan func(), 0)
	runWorkers(ctx, *parallel)

	pubsubCli := pubsub.NewPubSubClient(conn)
	log.Printf(`start subscriber to %s/%s(%s) with timeout %s \n`, *pubsubAddr, *topic, *timeout)
	log.Printf(`parallel =%d`, *parallel)
	job := func() {
		stream, err := pubsubCli.Subscribe(ctx, &pubsub.Topic{Name: *topic})
		if err != nil {
			log.Printf(`can't subscribe to topic %s: %s`, *topic, err)
		}
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				log.Printf(`exit`)
				return
			}
			if err != nil {
				log.Printf(`can't resv data: %s`, err)
				return
			}

			println(`recv`, atomic.AddInt64(&counter, 1))
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
