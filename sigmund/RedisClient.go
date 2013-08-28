package sigmund

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strings"
)

type RedisClient struct {
	Hostname string
	Prefix   string
	Conn     redis.Conn
	Psc      redis.PubSubConn
	Channels map[string]string
	Objects  map[string]string
}

func (r *RedisClient) Init() *RedisClient {

	//Parse the flags
	r.ParseFlags()

	//Member init
	r.Channels = make(map[string]string, 1)
	r.Objects = make(map[string]string, 1)

	r.Objects["hosts"] = r.Prefix + ":hosts" //List of Strings

	//Set up the format for where we're storing the Redis objects
	r.Objects["conn"] = r.Prefix + ":host:__hostname__:conn"       //String
	r.Objects["data"] = r.Prefix + ":host:__hostname__:data"       //String
	r.Objects["inputs"] = r.Prefix + ":host:__hostname__:inputs"   //Hash of Strings
	r.Objects["outputs"] = r.Prefix + ":host:__hostname__:outputs" //Hash of Strings

	//Set up the format for the Redis Pubsub channels
	r.Channels["hosts"] = r.Prefix + ":hosts"
	r.Channels["conn"] = r.Prefix + ":host:__hostname__"

	return r
}

/**
 * Helper getters for the objects and channels in Redis.
 */

func (r *RedisClient) Obj(name string, hostname string) string {
	return strings.Replace(r.Objects[name], "__hostname__", hostname, -1)
}

func (r *RedisClient) Chan(name string, hostname string) string {
	return strings.Replace(r.Channels[name], "__hostname__", hostname, -1)
}

/**
 * Initialisation functions
 */

func (r *RedisClient) ParseFlags() {
	flag.StringVar(&r.Prefix, "p", "sigmund", "Prefix set on all Sigmund Redis keys")
	flag.StringVar(&r.Hostname, "r", "localhost:6379", "Where to look for the Redis server, format `HOST:PORT`.")
}

func (r *RedisClient) Open() error {
	fmt.Println("Opening Redis connection.")

	//Return an error if we can't create the connection.
	conn, err := redis.Dial("tcp", r.Hostname)
	if err != nil {
		return err
	}

	//Assign the connection and a PubSub connection on top of the original.
	r.Conn = conn
	r.Psc = redis.PubSubConn{r.Conn}

	return nil
}

func (r *RedisClient) Close() {
	fmt.Println("Closing Redis connection.")
	r.Conn.Close()
}

/**
 * Redis helpers
 */

func (r *RedisClient) GetHosts() ([]string, error) {
	reply, err := r.Conn.Do("LRANGE", r.Obj("hosts", ""), 0, -1)

	if err != nil {
		return nil, err
	}

	return redis.Strings(reply, err)
}

func (r *RedisClient) HostExists(hostname string) (bool, error) {
	hosts, err := r.GetHosts()

	if err != nil {
		return false, err
	}

	exists := false
	for _, host := range hosts {
		if host == hostname {
			exists = true
			break
		}
	}

	return exists, nil
}

func (r *RedisClient) EnsureExists(hostname, string) error {
 		exists, err := r.HostExists(hostname)
 		if err != nil {
 			return err
 		}
 		if !exists {
 			return errors.New("Host" + hostname + " does not exist.")
 		}

 		return nil
}

/**
 * Adding a Host to Redis
 */

func (r *RedisClient) AddHost(hostname string) error {
	fmt.Println("Adding hostname")

	exists, err := r.HostExists(hostname)

	if err != nil || exists {
		return err
	}

	_, err = r.Conn.Do("RPUSH", r.Obj("hosts", ""), hostname)

	return err
}

func (r *RedisClient) AddConn(hostname string, data interface{}) error {
	fmt.Println("Adding conn")

	flatData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = r.Conn.Do("SET", r.Obj("conn", hostname), flatData)
	return err
}

func (r *RedisClient) AddData(hostname string, data interface{}) error {
	fmt.Println("Adding data")

	flatData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = r.Conn.Do("SET", r.Obj("data", hostname), flatData)
	return err
}

func (r *RedisClient) AddInputs(hostname string, data map[string]interface{}) error {
	fmt.Println("Adding inputs")

	for key, value := range data {
		flatValue, err := json.Marshal(value)
		if err != nil {
			return err
		}
		if _, err := r.Conn.Do("HSET", r.Obj("inputs", hostname), key, flatValue); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisClient) AddOutputs(hostname string, data map[string]interface{}) error {
	fmt.Println("Adding outputs")

	for key, value := range data {
		flatValue, err := json.Marshal(value)
		if err != nil {
			return err
		}
		if _, err := r.Conn.Do("HSET", r.Obj("outputs", hostname), key, flatValue); err != nil {
			return err
		}
	}
	return nil
}

/**
 * Getting a Host from Redis
 */

 func (r *RedisClient) GetConn(hostname string) string, error {
 		if err := r.EnsureExists(hostname); err != nil {
 			return err
 		}

 		return nil
 }

/**
 * Removing a Host from Redis
 */

func (r *RedisClient) RemoveHost(hostname string) error {
	_, err := r.Conn.Do("LREM", r.Obj("hosts", ""), 0, hostname)
	return err
}

func (r *RedisClient) RemoveConn(hostname string) error {
	_, err := r.Conn.Do("DEL", r.Obj("conn", ""), 0, hostname)
	return err
}

func (r *RedisClient) RemoveData(hostname string) error {
	_, err := r.Conn.Do("DEL", r.Obj("data", hostname))
	return err
}

func (r *RedisClient) RemoveInputs(hostname string) error {
	_, err := r.Conn.Do("DEL", r.Obj("inputs", hostname))
	return err
}

func (r *RedisClient) RemoveOutputs(hostname string) error {
	_, err := r.Conn.Do("DEL", r.Obj("outputs", hostname))
	return err
}
