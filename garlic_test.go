package garlic

import (
	"fmt"
	"testing"
	"time"

	"reflect"
)

var testPayload = []uint8{0x1, 0x0, 0x0, 0x0,
	0x1, 0x0, 0x0, 0x0,
	0xd5, 0x34, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0,
	0x28, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x80,
	0xb, 0x0, 0x0, 0x0,
	0x11, 0x13, 0x60, 0xff,
	0xaa, 0x72, 0x2, 0x0,
	0x99, 0x69, 0x0, 0x0,
	0xdd, 0x2d, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0,
	0xff, 0xff, 0xff, 0xff,
	0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0}

//These tests requires root. You've been warned.
// func TestDial(t *testing.T) {
// 	cn, err := DialPCN()
// 	if err != nil {
// 		t.Fatalf("dial error: %s", err)
// 	}

// 	for {
// 		data, err := cn.ReadPCN()
// 		//fmt.Printf("%#v\n", cn.Conn)
// 		if err != nil {
// 			t.Errorf("Test fail: %s", err)
// 		}
// 		fmt.Printf("%#v\n", data)
// 	}
// 	//fmt.Printf("%#v\n", data)

// 	cn.ClosePCN()
// }

// func TestDialFilter(t *testing.T) {
// 	cn, err := DialPCNWithEvents([]EventType{ProcEventGID, ProcEventExit})
// 	//cn, err := DialPCNWithEvent(ProcEventSID)
// 	if err != nil {
// 		t.Fatalf("%s", err)
// 	}

// 	for {
// 		data, err := cn.ReadPCN()
// 		//fmt.Printf("%#v\n", cn.Conn)
// 		if err != nil {
// 			t.Errorf("Test fail: %s", err)
// 		}
// 		fmt.Printf("%#v\n", data)
// 	}

// 	err = cn.RemoveBPF()
// 	if err != nil {
// 		t.Errorf("RemoveBPF: %s", err)
// 	}

// 	err = cn.ClosePCN()
// 	if err != nil {
// 		t.Errorf("ClosePCN: %s", err)
// 	}
// }

/*
This example demonstrates garlic in the most simplistic use case: connect and read
Proc connector requres root, and the underlying scoket returns an array from netlink
*/
func ExampleDialPCN() {
	cn, err := DialPCN()
	if err != nil {
		fmt.Printf("%s", err)
	}

	//Read in events
	for {
		data, err := cn.ReadPCN()

		if err != nil {
			fmt.Printf("Read fail: %s", err)
		}
		fmt.Printf("%#v\n", data)
	}
}

/*
This demonstrates reading a specific array of selected events
*/
func ExampleDialPCNWithEvents() {
	cn, err := DialPCNWithEvents([]EventType{ProcEventGID, ProcEventExit})
	if err != nil {
		fmt.Printf("%s", err)
	}

	//Read in events
	for {
		data, err := cn.ReadPCN()

		if err != nil {
			fmt.Printf("Read fail: %s", err)
		}
		fmt.Printf("%#v\n", data)
	}
}

func TestParseCN(t *testing.T) {

	validOut := ProcEvent{What: 0x80000000,
		CPU:        0xb,
		Timestamp:  time.Unix(0, 0x272aaff601311),
		WhatString: "Exit",
		EventData: Exit{ProcessPid: 0x6999,
			ProcessTgid: 0x2ddd,
			ExitCode:    0x0,
			ExitSignal:  0xffffffff}}

	connector := CnConn{c: nil, boottime: 0}

	ev, err := connector.parseCn(testPayload)
	if err != nil {
		t.Fatalf("Error parsing payload: %s", err)
	}

	if !reflect.DeepEqual(validOut, ev) {
		t.Fatalf("Events do not match: %#v \n %#v", ev, validOut)
	}

}

func TestIsACK(t *testing.T) {

	isNot := isAck(testPayload)
	if isNot {
		t.Fatalf("Packet is not an ACK %#v", testPayload)
	}

	ackPacket := []uint8{0x1, 0x0, 0x0, 0x0,
		0x1, 0x0, 0x0, 0x0,
		0x45, 0x28, 0x0, 0x0,
		0x1, 0x0, 0x0, 0x0,
		0x28, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x5, 0x0, 0x0, 0x0,
		0xeb, 0x63, 0x69, 0x90,
		0x69, 0x24, 0x2, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0}

	isAck := isAck(ackPacket)
	if !isAck {
		t.Fatalf("Pack is an ACK: %#v", ackPacket)
	}

}
