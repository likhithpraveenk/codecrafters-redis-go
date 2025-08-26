package store

import (
	"fmt"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/common"
)

type Client struct {
	Conn     net.Conn
	Messages chan []any
}

type PubSub struct {
	mu       sync.RWMutex
	channels map[string]map[*Client]struct{}
}

var pubsub = &PubSub{
	channels: make(map[string]map[*Client]struct{}),
}

var (
	clientsMu sync.RWMutex
	clients   = make(map[net.Conn]*Client)
)

func GetClient(conn net.Conn) *Client {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	if c, ok := clients[conn]; ok {
		return c
	}

	c := &Client{
		Conn:     conn,
		Messages: make(chan []any),
	}

	clients[conn] = c

	go func() {
		for msg := range c.Messages {
			_, err := conn.Write(common.Encode(msg))
			if err != nil {
				fmt.Println("Client disconnected:", err)
				return
			}
		}
	}()

	return c
}

func RemoveClient(conn net.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if c, ok := clients[conn]; ok {
		close(c.Messages)
		delete(clients, conn)
	}
}

func Subscribe(channel string, client *Client) []any {
	pubsub.mu.Lock()
	if pubsub.channels[channel] == nil {
		pubsub.channels[channel] = make(map[*Client]struct{})
	}
	pubsub.channels[channel][client] = struct{}{}
	pubsub.mu.Unlock()

	count := noOfSubscriptions(client)

	return []any{"subscribe", channel, count}
}

func noOfSubscriptions(client *Client) int64 {
	pubsub.mu.RLock()
	defer pubsub.mu.RUnlock()
	count := int64(0)
	for _, subscribers := range pubsub.channels {
		if _, ok := subscribers[client]; ok {
			count++
		}
	}
	return count
}

func Publish(channel, message string) int64 {
	pubsub.mu.RLock()
	clients := pubsub.channels[channel]
	pubsub.mu.RUnlock()

	for client := range clients {
		go func(c *Client) {
			c.Messages <- []any{"message", channel, message}
		}(client)
	}

	count := int64(len(pubsub.channels[channel]))

	return count
}

func UnSubscribe(channel string, client *Client) []any {
	pubsub.mu.Lock()
	defer pubsub.mu.Unlock()

	if clients, ok := pubsub.channels[channel]; ok {
		delete(clients, client)
	}

	count := noOfSubscriptions(client)

	return []any{"unsubscribe", channel, count}
}
