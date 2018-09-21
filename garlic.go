//Package garlic is a simple proc connector interface for golang.
package garlic

/*
GArLIC: GolAng LInux Connector: Linux Processor Connector library
*/

import (
	"encoding/binary"
	"fmt"
	"syscall"

	"github.com/mdlayher/netlink"
)

//parse and handle the event Interface
func getEvent(hdr procEventHdr, data []byte) (EventData, error) {
	switch hdr.What {
	case ProcEventNone:
		//We should only see this when we're getting an ACK back from the kernel
		return nil, fmt.Errorf("Got ProcEventNone")
	case ProcEventFork:
		ev := Fork{}
		ev.ParentPid, ev.ParentTgid, ev.ChildPid, ev.ChildTgid = return4Uint32(data)
		return ev, nil
	case ProcEventExec:
		ev := Exec{}
		ev.ProcessPid, ev.ProcessTgid = return2Uint32(data)
		return ev, nil
	case ProcEventUID, ProcEventGID:
		ev := ID{}
		ev.ProcessPid, ev.ProcessTgid, ev.RealID, ev.EffectiveID = return4Uint32(data)
		return ev, nil
	case ProcEventSID:
		ev := Sid{}
		ev.ProcessPid, ev.ProcessTgid = return2Uint32(data)
		return ev, nil
	case ProcEventPtrace:
		ev := Ptrace{}
		ev.ProcessPid, ev.ProcessTgid, ev.TracerPid, ev.TracerTgid = return4Uint32(data)
		return ev, nil
	case ProcEventComm:
		ev := Comm{}
		ev.ProcessPid, ev.ProcessTgid = return2Uint32(data)
		copy(ev.Comm[:], data[8:])
		return ev, nil
	case ProcEventCoredump:
		ev := Coredump{}
		ev.ProcessPid, ev.ProcessTgid = return2Uint32(data)
		return ev, nil
	case ProcEventExit:
		ev := Exit{}
		ev.ProcessPid, ev.ProcessTgid, ev.ExitCode, ev.ExitSignal = return4Uint32(data)
		return ev, nil
	}

	return Exit{}, fmt.Errorf("Unknown What: %x", hdr.What)
}

func parseCn(data []byte) (ProcEvent, error) {

	hdr := unmarshalProcEventHdr(data[cnMsgLen:])
	//buf := bytes.NewBuffer(data[cnMsgLen+procEventHdrLen:])

	ev, err := getEvent(hdr, data[cnMsgLen+procEventHdrLen:])
	if err != nil {
		return ProcEvent{}, err
	}

	return ProcEvent{What: hdr.What, CPU: hdr.CPU, TimestampNs: hdr.Timestamp, EventData: ev}, nil
}

//check to see if the packet is a valid ACK
func isAck(data []byte) bool {

	//buf := bytes.NewBuffer(data)
	msg := unmarshalCnMsg(data)
	hdr := unmarshalProcEventHdr(data[binary.Size(cnMsg{}):])

	if msg.Ack == 0x1 && hdr.What == ProcEventNone {
		return true
	}

	return false
}

//ClosePCN closes the netlink  connection
func (c CnConn) ClosePCN() error {
	return c.c.Close()
}

//ReadPCN reads waits for a Proc connector event to come across the nl socket, and returns an event struct
//This is a blocking operation
func (c CnConn) ReadPCN() ([]ProcEvent, error) {

	retMsg, err := c.c.Receive()
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

func dialPCN() (*netlink.Conn, error) {

	//DialPCN Config
	cCfg := netlink.Config{Groups: 0x1}

	//Bind
	c, err := netlink.Dial(syscall.NETLINK_CONNECTOR, &cCfg)
	//fmt.Println("Finished dial.")

	if err != nil {
		return &netlink.Conn{}, fmt.Errorf("Error in netlink: %s", err)
	}

	//setup process connector hdr
	cbHdr := cbID{Idx: CnIdxProc, Val: CnValProc}
	var connBody uint32 = ProcCnMcastListen
	cnHdr := cnMsg{ID: cbHdr, Len: uint16(binary.Size(connBody))}

	binHdr := cnHdr.marshalBinaryAndBody(connBody)

	reqMsg := netlink.Message{
		Header: netlink.Header{
			Type: syscall.NLMSG_DONE,
		},
		Data: binHdr,
	}

	//Send request message
	msgs, err := c.Send(reqMsg)
	if err != nil {
		return &netlink.Conn{}, fmt.Errorf("Execute error: %s\n %#v", err, msgs)
	}

	//Wait for our ack msg
	ack, err := c.Receive()
	if err != nil {
		return &netlink.Conn{}, fmt.Errorf("could not recv ack: %v", err)
	}

	//check to make sure out ack valid
	if !isAck(ack[0].Data) {
		return &netlink.Conn{}, fmt.Errorf("Packet not a valid ACK: %+v", ack)
	}

	return c, nil
}

//DialPCN connects to the proc connector socket, and returns a connection that will listens for all available event types:
//None, Fork, Execm UID, GID, SID, Ptrace, Comm, Coredump and Exit
func DialPCN() (CnConn, error) {

	c, err := dialPCN()

	return CnConn{c: c}, err

}

//DialPCNWithEvents is the same as DialPCN(), but with a filter that allows you select a particular proc event.
//It uses bitmasks and PBF to filter for the given events
func DialPCNWithEvents(events []EventType) (CnConn, error) {

	c, err := dialPCN()
	filters, err := loadBPF(events)
	if err != nil {
		return CnConn{}, err
	}
	err = c.SetBPF(filters)
	if err != nil {
		return CnConn{}, err
	}

	return CnConn{c: c}, nil

}
