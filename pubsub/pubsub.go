package pubsub

import (
	ppubsub "github.com/real-nil/grpc-pubsub/proto/pubsub"
	"golang.org/x/net/context"
	"google.golang.org/grpc/grpclog"
	"sync"
)

type Sender interface {
	Send(*ppubsub.Data) error
}

type server struct {
	lockForPubsuber *sync.RWMutex
	pubsubClinet    map[string][]Sender
}

func New() ppubsub.PubSubServer {
	return &server{
		lockForPubsuber: &sync.RWMutex{},
		pubsubClinet:    make(map[string][]Sender, 0),
	}
}

/*
        Subscribe(*Topic, PubSub_SubscribeServer) error
	Publish(context.Context, *Data) (*Answer, error)
*/

func (s *server) Subscribe(top *ppubsub.Topic, stream ppubsub.PubSub_SubscribeServer) error {
	grpclog.Printf("start subscribe for topic %s", top.Name)
	ctx := stream.Context()
	s.lockForPubsuber.Lock()
	s.pubsubClinet[top.Name] = append(s.pubsubClinet[top.Name], stream)
	s.lockForPubsuber.Unlock()
	// check context, return error if Done
	//stream, err := client.New(context.Background())
	//if err := nil{
	//	grpclog.Printf("incorrect create stream %s", stream)
	//	return errors.New("ololo")
	//}

	select {
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (s *server) Publish(ctx context.Context, data *ppubsub.Data) (*ppubsub.Answer, error) {
	var (
		senders []Sender
		ok      bool
	)
	s.lockForPubsuber.RLock()
	senders, ok = s.pubsubClinet[data.Topic]
	s.lockForPubsuber.RUnlock()

	if !ok {
		// OR send error return nil, errors.New()
		return &ppubsub.Answer{Subsriptions: 0}, nil
	}

	var c int32
	c = int32(len(senders))

	for _, stream := range senders {
		err := stream.Send(data)
		if err != nil {
			grpclog.Printf("incorrect streamt")
			c--
		}
	}

	return &ppubsub.Answer{Subsriptions: c}, nil
}
