package cps

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
)

var serverAddress = "crypta.io:8707"

// CPS (Centralized PubSub)
type CPS struct {
	conn   net.Conn
	mutex  sync.Mutex
	topics map[string][]*Subscription
}

type Subscription struct {
	c     *CPS
	ch    chan []byte
	topic string
}

func New() (*CPS, error) {
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		return nil, err
	}
	c := &CPS{
		conn:   conn,
		topics: make(map[string][]*Subscription),
	}
	go func() {
		r := bufio.NewReader(conn)
		for {
			line, _, err := r.ReadLine()
			if err != nil {
				fmt.Println(err)
				return
			}
			var obj []interface{}
			err = json.Unmarshal(line, &obj)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				continue
			}
			if len(obj) < 1 {
				continue
			}
			switch obj[0] {
			case "p":
				if len(obj) < 3 {
					continue
				}
				str, _ := obj[1].(string)
				dataStr, _ := obj[2].(string)
				data, err := base64.StdEncoding.DecodeString(dataStr)
				if err != nil {
					continue
				}
				c.mutex.Lock()
				for _, sub := range c.topics[str] {
					go func() {
						sub.ch <- data
					}()
				}
				c.mutex.Unlock()
			}
		}
	}()
	return c, nil
}

func (c *CPS) Subscribe(topic string) (*Subscription, error) {
	buf, err := json.Marshal([]interface{}{"s", topic})
	if err != nil {
		return nil, err
	}
	_, err = c.conn.Write(append(buf, '\n'))
	if err != nil {
		return nil, err
	}
	s := &Subscription{
		c:     c,
		ch:    make(chan []byte),
		topic: topic,
	}
	s.c.mutex.Lock()
	c.topics[topic] = append(c.topics[topic], s)
	s.c.mutex.Unlock()
	return s, nil
}

func (c *CPS) Publish(topic string, data []byte) error {
	buf, err := json.Marshal([]interface{}{"p", topic,
		base64.StdEncoding.EncodeToString(data),
	})
	if err != nil {
		return err
	}
	_, err = c.conn.Write(append(buf, '\n'))
	return err
}

func (s *Subscription) Topic() string {
	return s.topic
}

func (s *Subscription) Next(ctx context.Context) ([]byte, error) {
	select {
	case data := <-s.ch:
		return data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *Subscription) Cancel() {
	s.c.mutex.Lock()
	defer s.c.mutex.Unlock()
	for i := range s.c.topics[s.topic] {
		if s.c.topics[s.topic][i] == s {
			s.c.topics[s.topic] = append(
				s.c.topics[s.topic][:i],
				s.c.topics[s.topic][i+1:]...,
			)
			break
		}
	}
}
