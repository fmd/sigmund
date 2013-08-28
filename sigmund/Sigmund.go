package sigmund

/*
 * - Redis Struct represents the Redis connection.
 */

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"errors"
)

type Sigmund struct {
	Redis *RedisClient
}

func (b *Sigmund) Init() (*Sigmund, error) {

	//Set up Redis
	b.Redis = &RedisClient{}
	b.Redis.Init()

	if err := b.Redis.Open(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Sigmund) Shutdown() {
	b.Redis.Close()
}

func (b *Sigmund) CreateHost(path string) (*Host, error) {
	host := &Host{}
	host.Redis = b.Redis

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(file, host); err != nil {
		return nil, err
	}

	err, removeErr := host.Save()

	if removeErr != nil {
		return nil, removeErr
	}

	if err != nil {
		return nil, err
	}

	return host, nil
}

func (b *Sigmund) GetHost(hostname string) *Host, error {
	exists, err := b.Redis.HostExists()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("Host does not exist.")
	}

	host := &Host{}
	host.Redis = b.Redis

	

	return host, nil
}

func (b *Sigmund) RemoveHost(hostname string) error {
	host := b.GetHost(hostname)
	return host.Remove()
}
