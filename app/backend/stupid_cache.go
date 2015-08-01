package main

import (
	"errors"
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
		// TODO iterate over args, find any that match
		if args[2] == "EX" {
			time.AfterFunc(time.Duration(args[3].(int))*time.Second, func() {
				stupidCache[args[0].(string)] = nil
			})
		}
		// TODO NX
		return reply, nil
	}
	return reply, errors.New("Command " + commandName + "not supported by local cache")
}

func (conn stupidCacheConn) Send(commandName string, args ...interface{}) error {
	return errors.New("'Send' not supported by local cache")
}

func (conn stupidCacheConn) Flush() error {
	return nil
}

func (conn stupidCacheConn) Receive() (reply interface{}, err error) {
	return reply, errors.New("'Receive' not supported by local cache")
}

func checkIfKeyAllowed(key string, keyWhitelist []string) (err error) {
	for _, allowed := range keyWhitelist {
		if key == allowed {
			return nil
		}
	}
	return errors.New("Key " + " is not whitelisted for use in local cache")
}
