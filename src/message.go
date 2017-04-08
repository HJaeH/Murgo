package main


type Message struct {
	buf    []byte
	kind   uint16
	client *Client
}
