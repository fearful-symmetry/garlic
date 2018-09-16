package garlic

import "github.com/mdlayher/netlink/nlenc"

//convenience methods
func return4Uint32(data []byte) (uint32, uint32, uint32, uint32) {
	return nlenc.Uint32(data[0:4]),
		nlenc.Uint32(data[4:8]),
		nlenc.Uint32(data[8:12]),
		nlenc.Uint32(data[12:16])
}

func return2Uint32(data []byte) (uint32, uint32) {
	return nlenc.Uint32(data[0:4]),
		nlenc.Uint32(data[4:8])

}

/*
===============================================================================
These are the struct defs used in cn_proc.h

*/

//EventData is an interface that encapsulates the union type used in cn_proc
type EventData interface {
	Pid() uint32

	Tgid() uint32
}

//Fork is the event for process forks
type Fork struct {
	ParentPid  uint32
	ParentTgid uint32
	ChildPid   uint32
	ChildTgid  uint32
}

//Pid returns the event Process ID
func (f Fork) Pid() uint32 {
	return f.ChildPid
}

//Tgid returns the event thread group ID
func (f Fork) Tgid() uint32 {
	return f.ChildTgid
}

//Exec is the event for process exec()s
type Exec struct {
	ProcessPid  uint32
	ProcessTgid uint32
}

//Pid returns the event Process ID
func (e Exec) Pid() uint32 {
	return e.ProcessPid
}

//Tgid returns the event thread group ID
func (e Exec) Tgid() uint32 {
	return e.ProcessTgid
}

//ID represents UID/GID changes for a process.
//in cn_proc.h, the real/effective GID/UID is a series of union types, which Go does not have.
//creating a super-special interface for this would be overkill,
//So we're going to rename the vars and just use two.
//Consumers should use `what` to distinguish between the two.
type ID struct {
	ProcessPid  uint32
	ProcessTgid uint32
	RealID      uint32
	EffectiveID uint32
}

//Pid returns the event Process ID
func (i ID) Pid() uint32 {
	return i.ProcessPid
}

//Tgid returns the event thread group ID
func (i ID) Tgid() uint32 {
	return i.ProcessTgid
}

//Sid is the event for Session ID changes
type Sid struct {
	ProcessPid  uint32
	ProcessTgid uint32
}

//Pid returns the event process  ID
func (s Sid) Pid() uint32 {
	return s.ProcessPid
}

//Tgid returns the event thread group ID
func (s Sid) Tgid() uint32 {
	return s.ProcessTgid
}

//Ptrace is the event for ptrace events
type Ptrace struct {
	ProcessPid  uint32
	ProcessTgid uint32
	TracerPid   uint32
	TracerTgid  uint32
}

//Pid returns the event Process ID
func (p Ptrace) Pid() uint32 {
	return p.ProcessPid
}

//Tgid returns the event thread group ID
func (p Ptrace) Tgid() uint32 {
	return p.ProcessTgid
}

//Comm represents changes to the command name, /proc/$PID/comm
type Comm struct {
	ProcessPid  uint32
	ProcessTgid uint32
	Comm        [16]byte
}

//Pid returns the event Process ID
func (c Comm) Pid() uint32 {
	return c.ProcessPid
}

//Tgid returns the event thread group ID
func (c Comm) Tgid() uint32 {
	return c.ProcessTgid
}

//Coredump is the event for...core dumps
type Coredump struct {
	ProcessPid  uint32
	ProcessTgid uint32
}

//Pid returns the event Process ID
func (c Coredump) Pid() uint32 {
	return c.ProcessPid
}

//Tgid returns the event thread group ID
func (c Coredump) Tgid() uint32 {
	return c.ProcessTgid
}

//Exit is the event for exit()
type Exit struct {
	ProcessPid  uint32
	ProcessTgid uint32
	ExitCode    uint32
	ExitSignal  uint32
}

//Pid returns the event Process ID
func (e Exit) Pid() uint32 {
	return e.ProcessPid
}

//Tgid returns the event thread group ID
func (e Exit) Tgid() uint32 {
	return e.ProcessTgid
}
