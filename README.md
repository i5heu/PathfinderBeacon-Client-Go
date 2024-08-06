# PathfinderBeacon-Client-Go

This is the official Go client library for the [PathfinderBeacon Project](https://github.com/i5heu/PathfinderBeacon).  

## How to use
Check out the `cmd/example/main.go` file for a full example.

If not provided, the PathfinderBeacon-Client-Go will create a new private key which will lead to a new room.  
If you want to join an existing room, you need to provide the private key of the room.  
Anybody who has the room hash can view the room and its nodes.

```go
package main

import (
	"fmt"

	"github.com/i5heu/PathfinderBeacon-Client-Go"
	"github.com/i5heu/PathfinderBeacon/pkg/auth"
)

func main() {
	bootstraper, err := PathfinderBeacon.NewPathfinderBeacon(&auth.Key{})
	if err != nil {
		panic(err)
	}

	fmt.Println("privateKey", string(bootstraper.Key.PrivateKeyToPem()))

	err = bootstraper.AddIPsBestEffort(80, "tcp")
	if err != nil {
		fmt.Println("Could not add IPs: ", err)
	}
	// send the addresses to the server
	err = bootstraper.PushAddresses()
	if err != nil {
		panic(err)
	}

	room := bootstraper.GetRoomName()
	fmt.Println("roomName", room)

	err = bootstraper.PullNodes()
	if err != nil {
		panic(err)
	}

	nodes := bootstraper.GetNodes()
	for name, node := range nodes {
		fmt.Println("Node From Server ", name, " ->", node)
	}

	ok, err := auth.VerifyRoomSignature(room, bootstraper.GetRoomSignatureBase64(), bootstraper.GetPublicKeyBase64())
	if err != nil {
		panic(err)
	}
	fmt.Println("VerifyRoomSignature", ok)

	// you have to send bootstraper.PushAddresses once an hour, otherwise the node will be removed from the list
}
```