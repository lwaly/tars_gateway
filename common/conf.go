package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/goinggo/mapstructure"
)

var Conf *Config

// Config represents a configuration file.
type Config struct {
	filename string
	cache    *map[string]interface{}
}

// New creates a new Config object.
func NewConf(filename string) error {
	Conf = &Config{filename, nil}
	return Conf.Reload()
}

func InstanceGet() *Config {
	return Conf
}

func (c *Config) Bool(key string) (bool, error) {
	if err, value := c.getValue(key); err != nil {
		return false, err
	} else {
		if s, ok := value.(bool); ok {
			return s, nil
		} else {
			return false, errors.New(fmt.Sprintf("error key.%s", key))
		}
	}
}

func (c *Config) String(key string) (string, error) {
	if err, value := c.getValue(key); err != nil {
		return "", err
	} else {
		if s, ok := value.(string); ok {
			return s, nil
		} else {
			return "", errors.New(fmt.Sprintf("error key.%s", key))
		}
	}
}

// Float64 returns float64 type value.
func (c *Config) Float64(key string) (float64, error) {
	if err, value := c.getValue(key); err != nil {
		return 0, err
	} else {
		if s, ok := value.(float64); ok {
			return s, nil
		} else {
			return 0, errors.New(fmt.Sprintf("error key.%s", key))
		}
	}
}

// Int returns int type value.
func (c *Config) Int(key string) (int, error) {
	if err, value := c.getValue(key); err != nil {
		return 0, err
	} else {
		if s, ok := value.(float64); ok {
			return int(s), nil
		} else {
			return 0, errors.New(fmt.Sprintf("error key.%s", key))
		}
	}
}

// Int64 returns int64 type value.
func (c *Config) Int64(key string) (int64, error) {
	if err, value := c.getValue(key); err != nil {
		return 0, err
	} else {
		if s, ok := value.(float64); ok {
			return int64(s), nil
		} else {
			return 0, errors.New(fmt.Sprintf("error key.%s", key))
		}
	}
}

func (c *Config) GetStruct(key string, v interface{}) (err error) {
	err, value := c.getValue(key)
	if err != nil {
		return err
	}
	if err = mapstructure.Decode(value.(map[string]interface{}), v); err != nil {
		return
	}
	return
}

func (c *Config) GetArray(key string, v interface{}) (err error) {
	err, value := c.getValue(key)
	if err != nil {
		return err
	}
	if err = mapstructure.Decode(value.([]interface{}), v); err != nil {
		return
	}
	return
}

// getValue retreives a Config option into a passed in pointer or returns an error.
func (config *Config) getValue(key string) (err error, v interface{}) {
	if config.cache == nil {
		return errors.New("fail to read config"), nil
	}
	var ok bool
	if v, ok = (*config.cache)[key]; ok {
		if v == nil {
			return errors.New(fmt.Sprintf("error key.%s", key)), nil
		}
		return
	} else {
		return errors.New(fmt.Sprintf("error key.%s", key)), nil
	}

	return
}

// Reload clears the config cache.
func (config *Config) Reload() error {
	cache, err := primeCacheFromFile(config.filename)
	config.cache = cache

	if err != nil {
		return err
	}

	return nil
}

func (config *Config) Watch() {
	l := log.New(os.Stderr, "", 0)

	// Catch SIGHUP to automatically reload cache
	sighup := make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)

	for {
		<-sighup
		l.Println("Caught SIGHUP, reloading config...")
		config.Reload()
	}
}

func primeCacheFromFile(file string) (*map[string]interface{}, error) {
	// File exists?
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, err
	}

	// Read file
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Unmarshal
	var config map[string]interface{}
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
