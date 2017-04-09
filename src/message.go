package main

import(
)
type Message struct {
	buf    []byte
	kind   uint16
	client *Client
}
