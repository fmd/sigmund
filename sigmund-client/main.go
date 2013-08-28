package main

import (
	"flag"
	"github.com/fmd/sigmund/sigmund"
)

type Client struct {
	Hostname string
	IsDaemon bool
	Conn     *sigmund.Sigmund
}

func (c *Client) Init() {
	c.ParseFlags()
	c.Conn = &sigmund.Sigmund{}
	c.Conn.Init()
	flag.Parse()
}

func (c *Client) ParseFlags() {
	flag.BoolVar(&c.IsDaemon, "d", false, "Daemonize the client.")
	flag.StringVar(&c.Hostname, "h", "sigmund-client", "Hostname for the Sigmund client.")
}

func (c *Client) Serve() error {
	_, err := c.Conn.CreateHost("examples/flask/Bazfile")
	if err != nil {
		return err
	}

	return err
}

func main() {
	client := &Client{}
	client.Init()

	defer client.Conn.Shutdown()

	if err := client.Serve(); err != nil {
		panic(err)
	}

}
