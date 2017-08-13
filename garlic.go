//Package garlic is a simple proc connector interface for golang.
package garlic

/*
GArLIC: GolAng LInux Connector: Linux Processor Connector library
*/

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/mdlayher/netlink"
	"io"
	"syscall"
)

//from cn_proc.h
const (

	//ProcEventNone is only used for ACK events
	ProcEventNone = 0x00000000
	//ProcEventFork is a fork event
	ProcEventFork = 0x00000001
	//ProcEventExec is a exec() event
	ProcEventExec = 0x00000002
	//ProcEventUID is a user ID change
	ProcEventUID = 0x00000004
	//ProcEventGID is a group ID change
	ProcEventGID = 0x00000040
	//ProcEventSID is a session ID change
	ProcEventSID = 0x00000080
	//ProcEventSID is a process trace event
	ProcEventPtrace = 0x00000100
	//ProcEventComm is a comm(and) value change. Any value over 16 bytes will be truncated
	ProcEventComm = 0x00000200
	//ProcEventCoredump is a core dump event
	ProcEventCoredump = 0x40000000
	//ProcEventExit is an exit() event
	ProcEventExit = 0x80000000
)

//CnConn contains the connection to the proc connector socket
type CnConn struct {
	*netlink.Conn
}

//Various message structs from connector.h

type cbID struct {
	Idx uint32
	Val uint32
}

type cnMsg struct {
	ID    cbID
	Seq   uint32
	Ack   uint32
	Len   uint16
	Flags uint16
}

//This is just an internal  header that allows us to easily cast the raw binary data
type procEventHdr struct {
	What      uint32
	CPU       uint32
	Timestamp uint64
}

//parse and handle the event Interface
func getEvent(hdr procEventHdr, buf io.Reader) (EventData, error) {
	switch hdr.What {
	case ProcEventNone:
		//We should only see this when we're getting an ACK back from the kernel
		return nil, fmt.Errorf("Got ProcEventNone")
	case ProcEventFork:
		ev := Fork{}
		err := binary.Read(buf, binary.LittleEndian, &ev)
		return ev, err
	case ProcEventExec:
		ev := Exec{}
		err := binary.Read(buf, binary.LittleEndian, &ev)
		return ev, err
	case ProcEventUID, ProcEventGID:
		ev := ID{}
		err := binary.Read(buf, binary.LittleEndian, &ev)
		return ev, err
	case ProcEventSID:
		ev := Sid{}
		err := binary.Read(buf, binary.LittleEndian, &ev)
		return ev, err
	case ProcEventPtrace:
		ev := Ptrace{}
		err := binary.Read(buf, binary.LittleEndian, &ev)
		return ev, err
	case ProcEventComm:
		ev := Comm{}
		err := binary.Read(buf, binary.LittleEndian, &ev)
		return ev, err
	case ProcEventCoredump:
		ev := Coredump{}
		err := binary.Read(buf, binary.LittleEndian, &ev)
		return ev, err
	case ProcEventExit:
		ev := Exit{}
		err := binary.Read(buf, binary.LittleEndian, &ev)
		return ev, err
	}

	return Exit{}, fmt.Errorf("Unknown What: %x", hdr.What)
}

func getHeaders(buf io.Reader) (cnMsg, procEventHdr, error) {
	msg := cnMsg{}
	hdr := procEventHdr{}

	err := binary.Read(buf, binary.LittleEndian, &msg)
	if err != nil {
		return msg, hdr, err
	}

	err = binary.Read(buf, binary.LittleEndian, &hdr)

	return msg, hdr, err
}

func parseCn(data []byte) (ProcEvent, error) {

	buf := bytes.NewBuffer(data)
	_, hdr, err := getHeaders(buf)
	if err != nil {
		return ProcEvent{}, err
	}

	ev, err := getEvent(hdr, buf)
	if err != nil {
		return ProcEvent{}, err
	}

	return ProcEvent{What: hdr.What, CPU: hdr.CPU, TimestampNs: hdr.Timestamp, EventData: ev}, nil
}

//check to see if the packet is a valid ACK
func isAck(data []byte) bool {

	buf := bytes.NewBuffer(data)
	msg, hdr, err := getHeaders(buf)
	if err != nil {
		return false
	}

	if msg.Ack == 0x1 && hdr.What == ProcEventNone {
		return true
	}

	return false
}

//ClosePCN closes the netlink  connection
func (c CnConn) ClosePCN() error {
	return c.Close()
}

//ReadPCN reads waits for a Proc connector event to come across the nl socket, returns an event struct
func (c CnConn) ReadPCN() ([]ProcEvent, error) {

	retMsg, err := c.Receive()
	if err != nil {
		return nil, fmt.Errorf("Receive error: %s", err)
	}

	//I've never seen these underlying libs return more than one proc event, but lets not make assumptions
	evList := make([]ProcEvent, len(retMsg))
	for iter, value := range retMsg {
		parsedEv, err := parseCn(value.Data)
		if err != nil {
			return nil, fmt.Errorf("Bad parseCn: %s", err)
		}
		evList[iter] = parsedEv

	}

	return evList, nil
}

//DialPCN connects to the proc connector socket
func DialPCN() (CnConn, error) {

	//DialPCN Config
	cCfg := netlink.Config{Groups: 0x1}

	//Bind
	c, err := netlink.Dial(syscall.NETLINK_CONNECTOR, &cCfg)
	//fmt.Println("Finished dial.")

	if err != nil {
		return CnConn{}, fmt.Errorf("Error in netlink.DialPCN: %s", err)
	}

	//setup process connector hdr
	cbHdr := cbID{Idx: 0x1, Val: 0x1}
	var connBody uint32 = 0x1
	cnHdr := cnMsg{ID: cbHdr, Len: uint16(binary.Size(connBody))}

	buf := bytes.NewBuffer(make([]byte, 0, binary.Size(cnHdr)+binary.Size(connBody)))
	err = binary.Write(buf, binary.LittleEndian, cnHdr)

	if err != nil {
		return CnConn{}, err
	}

	err = binary.Write(buf, binary.LittleEndian, connBody)

	if err != nil {
		return CnConn{}, err
	}

	reqMsg := netlink.Message{
		Header: netlink.Header{
			Type: syscall.NLMSG_DONE,
		},
		Data: buf.Bytes(),
	}

	//Send request message
	msgs, err := c.Send(reqMsg)
	if err != nil {
		return CnConn{}, fmt.Errorf("Execute error: %s\n %#v", err, msgs)
	}

	//Wait for our ack msg
	ack, err := c.Receive()
	if err != nil {
		return CnConn{}, err
	}

	//check to make sure out ack valid
	if !isAck(ack[0].Data) {
		return CnConn{}, fmt.Errorf("Packet not a valid ACK: %+v", ack)
	}

	return CnConn{c}, nil

}
