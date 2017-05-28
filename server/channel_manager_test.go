package server

import (
	"fmt"
	"murgo/pkg/moduleserver"
)

type CMTest1 struct {
}

func (cmTest *CMTest1) startLink() {
	fmt.Print("startlink")
}

func (cmTest *CMTest1) init() {
	msg := &moduleserver.CastMessage{}
	moduleserver.genServer.Cast("CM2", msg)
}

func (cmTest *CMTest1) terminate() {
	fmt.Print("startlink")
}

func (cmTest *CMTest1) handleCall(msg *moduleserver.CallMessage) {
	fmt.Println(msg, "ddd")
}

func (cmTest *CMTest1) handleCast(msg *moduleserver.CastMessage) {
	fmt.Println(msg, "ddd")
}

type CMTest2 struct {
}

func (cmTest *CMTest2) startLink() {
	fmt.Print("startlink")
}

func (cmTest *CMTest2) init() {
	msg := &moduleserver.CastMessage{}
	moduleserver.genServer.Cast("CM1", msg)
}

func (cmTest *CMTest2) terminate() {
	fmt.Print("startlink")
}

func (cmTest *CMTest2) handleCall() {
	fmt.Print("startlink")
}

func (cmTest *CMTest2) handleCast(msg *moduleserver.CastMessage) {
	fmt.Println(msg, "ddd")
}
