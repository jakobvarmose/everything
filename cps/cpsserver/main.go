package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

func main() {
	s, err := net.Listen("tcp", ":8707")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	for {
		c, err := s.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}
		go handle(c)
	}
}

var subscriptions sync.Map

func handle(c net.Conn) {
	defer c.Close()
	r := bufio.NewReader(c)
	for {
		line, _, err := r.ReadLine()
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
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
		case "s":
			if len(obj) < 2 {
				continue
			}
			str, _ := obj[1].(string)
			m, _ := subscriptions.LoadOrStore(str, new(sync.Map))
			m.(*sync.Map).Store(c, true)
		case "u":
			if len(obj) < 2 {
				continue
			}
			str, _ := obj[1].(string)
			m, _ := subscriptions.LoadOrStore(str, new(sync.Map))
			m.(*sync.Map).Delete(c)
		case "p":
			if len(obj) < 3 {
				continue
			}
			line2, err := json.Marshal(obj)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				continue
			}
			str, _ := obj[1].(string)
			m, _ := subscriptions.LoadOrStore(str, new(sync.Map))
			m.(*sync.Map).Range(func(key interface{}, value interface{}) bool {
				key.(net.Conn).Write(append(line2, '\n'))
				return true
			})
		case "x":
			return
		}
	}
}
