package main

import (
	"errors"
	"fmt"
	"time"
)

var stupidCache map[string][]byte

type stupidCacheConn struct {
	AllowedKeys []string
}

func (conn stupidCacheConn) Close() error {
	return nil
}

func (conn stupidCacheConn) Err() error {
	return nil
}

func (conn stupidCacheConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if stupidCache == nil {
		stupidCache = make(map[string][]byte)
	}
	if commandName == "GET" {
		if err := checkIfKeyAllowed(args[0].(string), conn.AllowedKeys); err != nil {
			return nil, err
		}
		return stupidCache[args[0].(string)], nil
	} else if commandName == "SET" {
		if err := checkIfKeyAllowed(args[0].(string), conn.AllowedKeys); err != nil {
			return nil, err
		}
		stupidCache[args[0].(string)] = args[1].([]byte)
		parseEX := false
		for _, arg := range args[2:] {
			if arg == "EX" {
				parseEX = true
			} else if parseEX {
				parseEX = false
				time.AfterFunc(time.Duration(arg.(int))*time.Second, func() {
					stupidCache[args[0].(string)] = nil
				})
			} else if arg == "NX" {
				// noop
			} else {
				return reply, fmt.Errorf("Option %v is not supported by local cache", arg)
			}
		}
		return reply, nil
	}
	return reply, fmt.Errorf("Command %v is not supported by local cache", commandName)
}

func (conn stupidCacheConn) Send(commandName string, args ...interface{}) error {
	return errors.New("'Send' is not supported by local cache")
}

func (conn stupidCacheConn) Flush() error {
	return nil
}

func (conn stupidCacheConn) Receive() (reply interface{}, err error) {
	return reply, errors.New("'Receive' is not supported by local cache")
}

func checkIfKeyAllowed(key string, keyWhitelist []string) (err error) {
	for _, allowed := range keyWhitelist {
		if key == allowed {
			return nil
		}
	}
	return fmt.Errorf("Key %v is not whitelisted for use in local cache", key)
}
