package main


type Channel struct {
	Id       int
	parent    *Channel
	Name     string
	Links map[int]*Channel
	Description string
	Position int
}

