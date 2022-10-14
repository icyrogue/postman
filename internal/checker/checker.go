package checker

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

type checker struct {
	client  *client.Client
	storage Storage
	Options *Options
}

type Options struct {
	Server string
	Port   string

	Login    string
	Password string

	WaitTime int
}

type Storage interface {
	AddRead(ctx context.Context, tmStamp, addr string) error
}

func New(storage Storage) *checker {
	return &checker{storage: storage}
}

func (c *checker) Init() error {
	client, err := client.DialTLS(c.Options.Server+":"+c.Options.Port, nil)
	if err != nil {
		return err
	}

	if err := client.Login(c.Options.Login, c.Options.Password); err != nil {
		return err
	}
	c.client = client
	return nil
}

func (c *checker) Close() error {
	return c.client.Close()
}

func (c *checker) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Second * time.Duration(c.Options.WaitTime))
	inbox, err := c.client.Select("INBOX", false)
	if err != nil {
		log.Println(err.Error())
	}
	var prevCount = uint32(inbox.Messages)
	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-ticker.C:
				log.Println("got an update")
				inbox, err := c.client.Select("INBOX", false)
				if err != nil {
					log.Println(err.Error())
				}
				var count uint32
				if count = inbox.Messages; count == prevCount {
					log.Println("not this time")
					continue loop
				}
				err = c.check(prevCount, count)
				if err != nil {
					log.Println(err.Error())
				}
				prevCount = count
			}
		}
		//graceful shutdown routine
	}()
}

func (c *checker) check(from, count uint32) error {
	log.Println("Checking!", count, from)
	var section imap.BodySectionName
	section.Specifier = imap.HeaderSpecifier

	seqset := new(imap.SeqSet)
	seqset.AddRange(from, count)
	//Del
	log.Println(seqset)
	messages := make(chan *imap.Message, 55)
	done := make(chan error, 1)
	go func() {
		done <- c.client.Fetch(seqset, []imap.FetchItem{section.FetchItem()}, messages)
	}()
	if err := <-done; err != nil {
		return err
	}
	time.Sleep(time.Second * 10)
	log.Println(len(messages))
	for msg := range messages {
		data, err := mail.CreateReader(msg.GetBody(&section))
		if err != nil {
			return err
		}
		_, params, err := data.Header.ContentType()
		if err != nil {
			return err
		}
		hd := params["report-type"]
		if strings.Contains(hd, "disposition-notification") {
			tmStamp, err := data.Header.Date()
			if err != nil {
				log.Println(err.Error())
				continue
			}
			addr, err := data.Header.AddressList("Reply-To")
			if err != nil {
				log.Println(err.Error())
				continue
			}
			if err = c.storage.AddRead(context.Background(), addr[0].String(), tmStamp.Format(time.RFC1123)); err != nil {
				log.Println(err.Error())
				continue
			}
		}
	}
	return nil
}
