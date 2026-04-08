package ws

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

// ChannelLayer defines the interface for cross-process/cross-goroutine WebSocket communication.
type ChannelLayer interface {
	Send(channel string, message any) error
	Receive(channel string) (any, error)
	GroupAdd(group string, channel string) error
	GroupDiscard(group string, channel string) error
	GroupSend(group string, message any) error
}

// InMemoryChannelLayer uses local channels and maps. Only suitable for single-process setups.
type InMemoryChannelLayer struct {
	channels sync.Map // map[string]chan any
	groups   sync.Map // map[string]map[string]bool
}

// NewInMemoryChannelLayer creates a new local channel layer.
func NewInMemoryChannelLayer() *InMemoryChannelLayer {
	return &InMemoryChannelLayer{}
}

func (l *InMemoryChannelLayer) getChannel(name string) chan any {
	ch, _ := l.channels.LoadOrStore(name, make(chan any, 100))
	return ch.(chan any)
}

func (l *InMemoryChannelLayer) Send(channel string, message any) error {
	ch := l.getChannel(channel)
	ch <- message
	return nil
}

func (l *InMemoryChannelLayer) Receive(channel string) (any, error) {
	ch := l.getChannel(channel)
	msg := <-ch
	return msg, nil
}

func (l *InMemoryChannelLayer) GroupAdd(group string, channel string) error {
	var members sync.Map
	actual, _ := l.groups.LoadOrStore(group, &members)
	actualMembers := actual.(*sync.Map)
	actualMembers.Store(channel, true)
	return nil
}

func (l *InMemoryChannelLayer) GroupDiscard(group string, channel string) error {
	if actual, ok := l.groups.Load(group); ok {
		members := actual.(*sync.Map)
		members.Delete(channel)
	}
	return nil
}

func (l *InMemoryChannelLayer) GroupSend(group string, message any) error {
	if actual, ok := l.groups.Load(group); ok {
		members := actual.(*sync.Map)
		members.Range(func(key, value any) bool {
			channel := key.(string)
			l.Send(channel, message)
			return true
		})
	}
	return nil
}


// RedisChannelLayer uses Redis pub/sub for multi-process scalability.
type RedisChannelLayer struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisChannelLayer creates a new Redis channel layer.
func NewRedisChannelLayer(addr string) *RedisChannelLayer {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisChannelLayer{
		client: rdb,
		ctx:    context.Background(),
	}
}

func (l *RedisChannelLayer) Send(channel string, message any) error {
	// message should be JSON encodable
	return l.client.Publish(l.ctx, channel, message).Err()
}

func (l *RedisChannelLayer) Receive(channel string) (any, error) {
	pubsub := l.client.Subscribe(l.ctx, channel)
	defer pubsub.Close()

	msg, err := pubsub.ReceiveMessage(l.ctx)
	if err != nil {
		return nil, err
	}
	return msg.Payload, nil
}

func (l *RedisChannelLayer) GroupAdd(group string, channel string) error {
	// In Redis, groups can be implemented using Sets
	return l.client.SAdd(l.ctx, "group:"+group, channel).Err()
}

func (l *RedisChannelLayer) GroupDiscard(group string, channel string) error {
	return l.client.SRem(l.ctx, "group:"+group, channel).Err()
}

func (l *RedisChannelLayer) GroupSend(group string, message any) error {
	members, err := l.client.SMembers(l.ctx, "group:"+group).Result()
	if err != nil {
		return err
	}

	var errs []error
	for _, member := range members {
		if err := l.Send(member, message); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("group send errors: %v", errs)
	}
	return nil
}
