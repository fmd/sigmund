package main

import (
	"fmt"
	"flag"
	"os/signal"
	"os"
	"syscall"
	"github.com/fmd/sigmund/sigmund"
	"runtime"
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

func (c *Client) GetMessages() <-chan bool {
	return make(chan bool, 1)
}

func (c *Client) GetInterrupt() <-chan bool {
    sigChan  := make(chan os.Signal, 1)
    killChan := make(chan bool, 1)

    signal.Notify(sigChan, os.Interrupt)
    signal.Notify(sigChan, syscall.SIGTERM)

    go func() {
        <-sigChan
        fmt.Println("COOL")
        killChan <- true
    }()

    return killChan
}

func (c *Client) ServeOnce() <-chan bool {
 return make(chan bool, 1)
}

func (c *Client) Serve() error {
	_, err := c.Conn.CreateHost("../examples/flask/Sigfile") 
	if err != nil {
		return err
	}

	interrupted := c.GetInterrupt()
	i := 0
	ib := false
	for ib != true {
		i++
		select {
			case <-interrupted:
				fmt.Println("Interrupted!")
				ib = true
			default:
		}
		runtime.Gosched()
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
