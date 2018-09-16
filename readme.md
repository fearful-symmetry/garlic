# GArLIC: GolAng LInux Connector

GARLIC is a simple proc connector interface for golang.

It's dead simple, and built on top of @mdlayher's [gloang netlink library](https://github.com/mdlayher/netlink)

The Proc Connector interface is mildly obscure, and you can read more [here](http://netsplit.com/the-proc-connector-and-socket-filters)

[![Go Report Card](https://goreportcard.com/badge/github.com/fearful-symmetry/garlic)](https://goreportcard.com/report/github.com/fearful-symmetry/garlic)
[![CircleCI](https://circleci.com/gh/fearful-symmetry/garlic.svg?style=svg)](https://circleci.com/gh/fearful-symmetry/garlic)
## Tutorial

```go
//Open a connection to the local Proc connector instance
//This requires root.
cn, err := DialPCN()
	if err != nil {
		log.Fatalf("%s", err)
	}

//Read in events
for {
    data, err = cn.ReadPCN()

	if err != nil {
		log.Errorf("Read fail: %s", err)
    }
	fmt.Printf("%#v\n", data)
}

//You can also filter by a list of events
cn, err := DialPCNWithEvents([]EventType{ProcEventGID, ProcEventExit})
	if err != nil {
		log.Fatalf("%s", err)
	}

```

## Why?

Because it's fun. Also, garlic is my favorite seasoning.

## What's next?

- A CLI implementation is in the works.
- Find a non-root way to run the tests.
- Start looking at perf data
- add new interfaces
- Expand tests
