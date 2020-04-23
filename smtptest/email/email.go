// Copyright 2018 Jakob Varmose Bentzen
package email

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"net/smtp"
	"strings"
	"time"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
}

type Message struct {
	Sender  string
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Body    string
}

func Send(cfg *Config, msg *Message) error {
	recipients := append(msg.To, msg.Cc...)
	recipients = append(recipients, msg.Bcc...)

	if strings.ContainsAny(msg.Sender, "\r\n") {
		return errors.New("The sender email address  must not contain CR or LF")
	}
	for _, recipient := range recipients {
		if strings.ContainsAny(recipient, "\r\n") {
			return errors.New("A recipient email address must not contain CR or LF")
		}
	}
	if strings.ContainsAny(msg.Subject, "\r\n") {
		return errors.New("The subject must not contain CR or LF")
	}

	c, err := smtp.Dial(cfg.Host + ":" + cfg.Port)
	if err != nil {
		return err
	}
	defer c.Quit()

	err = c.StartTLS(&tls.Config{
		ServerName: cfg.Host,
	})
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	err = c.Auth(auth)
	if err != nil {
		return err
	}

	err = c.Mail(msg.Sender)
	if err != nil {
		return err
	}

	for _, recipient := range recipients {
		err := c.Rcpt(recipient)
		if err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	idBuf := make([]byte, 16)
	_, err = rand.Read(idBuf)
	if err != nil {
		return err
	}

	id := "<" + hex.EncodeToString(idBuf) + "." + msg.Sender + ">"
	_, err = w.Write([]byte("Message-ID: " + id + "\r\n"))
	if err != nil {
		return err
	}

	date := time.Now().UTC().Format(time.RFC1123Z)
	_, err = w.Write([]byte("Date: " + date + "\r\n"))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte("From: " + msg.Sender + "\r\n"))
	if err != nil {
		return err
	}

	for _, to := range msg.To {
		_, err = w.Write([]byte("To: " + to + "\r\n"))
		if err != nil {
			return err
		}
	}

	for _, cc := range msg.Cc {
		_, err = w.Write([]byte("Cc: " + cc + "\r\n"))
		if err != nil {
			return err
		}
	}

	_, err = w.Write([]byte("Subject: " + msg.Subject + "\r\n\r\n"))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(msg.Body))
	if err != nil {
		return err
	}

	return w.Close()
}
