package main

import (
	"flag"
	"fmt"
	"github.com/fmd/sigmund/sigmund"
	"github.com/garyburd/redigo/redis"
)

/*

	HOW BAZAAR SHOULD WORK:

	-- One Docker Host, 3 Containers --

	Container A:
		Inputs:  "mysql", "mongodb"
		Outputs: ""

	Container B:
		Inputs:  "mysql"
		Outputs: "mongodb"

	Container C:
		Inputs:  ""
		Outputs: "mysql"


	* Sigmund server starts on host.
	* Sigmund server subscribes to "sigmund:hosts"

	* Host starts container A.
	* Sigmund client starts in container A.
	* Sigmund client subscribes to "sigmund:host:A"
	* Sigmund client sets its info to sigmund:host:A, sigmund:host:A:inputs and sigmund:host:A:outputs
	* Sigmund client in container A connects to redis and publishes "new" to sigmund:hosts.

	* Sigmund server sees the published message.
	* Sigmund server creates the network bridge if it does not exist
	* Sigmund server looks at sigmund:host:A:inputs, and checks if they match to any other hosts' outputs. Nope.
	* Sigmund server looks at sigmund:host:A:outputs, and checks if they match to any other hosts' inputs. Nope.

	* Host starts container B.
	* Sigmund client starts in container B.
	* Sigmund client subscribes to "sigmund:host:B"
	* Sigmund client sets its info to sigmund:host:B, sigmund:host:B:inputs and sigmund:host:B:outputs
	* Sigmund client in container B connects to redis and publishes "new" to sigmund:hosts.

	* Sigmund server sees the published message.
	* Sigmund server looks at sigmund:host:B:inputs, and checks if they match to any other hosts' outputs. Nope.
	* Sigmund server looks at sigmund:host:B:outputs, and checks if they match to any other hosts' inputs. Yes!
	* Sigmund server sends "inputs:mongodb host:B:outputs:mongodb" to "sigmund:host:A"

	* Sigmund client on host A receives the message.
	* Sigmund client on host A finds and parses the output.
	* Sigmund client on host A compiles .sig files and executes post compile

	* Host starts container C
	* Sigmund client starts in container C
	* Sigmund client sets its info to sigmund:host:C, sigmund:host:C:inputs and sigmund:host:C:outputs
	* Sigmund client in container C connects to redis and publishes "new" to sigmund:hosts.

	* Sigmund server sees the published message.
	* Sigmund server looks at sigmund:host:C:inputs, and checks if they match to any other hosts' outputs. Nope.
	* Sigmund server looks at sigmund:host:C:outputs, and checks if they match to any other hosts' inputs. Yes!
	* Sigmund server sends "inputs:mysql host:C:outputs:mysql" to "sigmund:host:A"

	* Sigmund client on host A receives the message.
	* Sigmund client on host A finds and parses the output.
	* Sigmund client on host A compiles .sig files and executes post compile

*/

/*
 * - Server Struct represents the Sigmund server.
 */

type Server struct {
	IsDaemon bool
	Bridge   string
	Conn     *sigmund.Sigmund
}

func (s *Server) Init() *Server {
	s.ParseFlags()
	s.Conn = &sigmund.Sigmund{}
	s.Conn.Init()
	flag.Parse()

	return s
}

func (s *Server) ParseFlags() {
	flag.BoolVar(&s.IsDaemon, "d", false, "Daemonize the server.")
	flag.StringVar(&s.Bridge, "i", "baz0", "Name of the network bridge to create for Sigmund.")
}

func (s *Server) Serve() error {
	for {
		switch n := s.Conn.Redis.Psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("Message: %s %s\n", n.Channel, n.Data)
		case redis.Subscription:
			fmt.Printf("Subscription: %s %s %d\n", n.Kind, n.Channel, n.Count)
			if n.Count == 0 {
				return nil
			}
		case error:
			fmt.Printf("error: %v\n", n)
			return n
		}
	}

	return nil
}

/*
 * - Main func starts the program
 */

func main() {
	server := &Server{}
	server.Init()

	if err := server.Conn.Redis.Open(); err != nil {
		panic(err)
	}

	defer server.Conn.Redis.Close()

	if err := server.Serve(); err != nil {
		panic(err)
	}
}
