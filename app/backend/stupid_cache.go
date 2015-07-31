package main

import (
	"errors"
	"time"
)

var stupidCache map[string][]byte

type stupidCacheConn struct{}

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
		return stupidCache[args[0].(string)], nil
	} else if commandName == "SET" {
		stupidCache[args[0].(string)] = args[1].([]byte)
		if args[2] == "EX" {
			time.AfterFunc(time.Duration(args[3].(int))*time.Second, func() {
				stupidCache[args[0].(string)] = nil
			})
		}
		return reply, nil
	}
	return reply, errors.New("Command not yet implemented")
}

func (conn stupidCacheConn) Send(commandName string, args ...interface{}) error {
	return errors.New("Not yet implemented")
}

func (conn stupidCacheConn) Flush() error {
	return nil
}

func (conn stupidCacheConn) Receive() (reply interface{}, err error) {
	return reply, errors.New("Not yet implemented")
}
