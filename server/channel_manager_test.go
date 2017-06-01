package server

import (
	"fmt"
	"murgo/pkg/servermodule"
)

type CMTest1 struct {
}

func (cmTest *CMTest1) startLink() {
	fmt.Print("startlink")
}

func (cmTest *CMTest1) init() {
	msg := &servermodule.CastMessage{}
	servermodule.genServer.Cast("CM2", msg)
}

func (cmTest *CMTest1) terminate() {
	fmt.Print("startlink")
}

func (cmTest *CMTest1) handleCall(msg *servermodule.CallMessage) {
	fmt.Println(msg, "ddd")
}

func (cmTest *CMTest1) handleCast(msg *servermodule.CastMessage) {
	fmt.Println(msg, "ddd")
}

type CMTest2 struct {
}

func (cmTest *CMTest2) startLink() {
	fmt.Print("startlink")
}

func (cmTest *CMTest2) init() {
	msg := &servermodule.CastMessage{}
	servermodule.genServer.Cast("CM1", msg)
}

func (cmTest *CMTest2) terminate() {
	fmt.Print("startlink")
}

func (cmTest *CMTest2) handleCall() {
	fmt.Print("startlink")
}

func (cmTest *CMTest2) handleCast(msg *servermodule.CastMessage) {
	fmt.Println(msg, "ddd")
}
