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

	req, err := bootstraper.GetPushAddressesRaw()
	if err != nil {
		panic(err)
	}
	fmt.Println("Request:", req)

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
