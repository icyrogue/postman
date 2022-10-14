package client

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

type client struct {
	client  *smtp.Client
	Options *Options
}

type Options struct {
	Port   string
	Server string

	Password string
	Login    string
	Name     string
}

func New() *client {
	return &client{}
}

//Init: подключается к smtp, авторизируется
func (c *client) Init() error {
	client, err := smtp.Dial(c.Options.Server + ":" + c.Options.Port)
	if err != nil {
		return err
	}
	if err := client.StartTLS(&tls.Config{}); err != nil {
		return err
	}
	auth := sasl.NewPlainClient("", c.Options.Login, c.Options.Password)
	if err := client.Auth(auth); err != nil {
		return err
	}

	if err != nil {
		return err
	}
	c.client = client
	return nil
}

//Close: аккуртано отключаемся
func (c *client) Close() error {
	if err := c.client.Quit(); err != nil {
		return err
	}
	return c.client.Close()
}

//Send: отправляет email на адрес addr из списка c id, c телом body в формате text/html
func (c *client) Send(id, addr string, body []byte) error {
	log.Println("preparing to send email from mailing list", id)
	if err := c.client.Mail(c.Options.Login, nil); err != nil {
		return err
	}
	if err := c.client.Rcpt(addr); err != nil {
		return err
	}
	wc, err := c.client.Data()
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintf(wc, "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\nDisposition-Notification-To: %v\nTo: %v\n", c.Options.Login, addr); err != nil {
		return err
	}
	if _, err = wc.Write(body); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	log.Println("email sent to", addr)
	return nil
}
