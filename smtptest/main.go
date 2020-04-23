// Copyright 2018 Jakob Varmose Bentzen
package main

import (
	"bufio"
	"fmt"
	"jakobvarmose/smtptest/email"
	"net/textproto"
	"os"
)

func main() {
	cfg := &email.Config{
		Host:     "smtp.mailtrap.io",
		Port:     "2525",
		Username: "",
		Password: "",
	}

	msg := &email.Message{
		Sender:  "test@1.com",
		To:      []string{"test@2.com"},
		Subject: "ABC",
		Body:    "Hello world. This is a test",
	}

	err := email.Send(cfg, msg)
	if err != nil {
		fmt.Println(err)
	}
}
