package main

import (
	"fmt"

	"github.com/i5heu/PathfinderBeacon-Client-Go"
	"github.com/i5heu/PathfinderBeacon/pkg/auth"
	"golang.org/x/exp/rand"
)

func main() {
	bootstraper, err := PathfinderBeacon.NewPathfinderBeacon(&auth.Key{})
	if err != nil {
		panic(err)
	}

	fmt.Println("privateKey", string(bootstraper.Key.PrivateKeyToPem()))

	bootstraper.AddAddress("127.0.0.1", 53, "udp")
	bootstraper.AddAddress("10.0.0.0", 80, "tcp")
	bootstraper.AddAddress("10.0.0.1", 443, "tcp")
	bootstraper.AddAddress("192.168.0.6", 443, "tcp")

	// add 40 random addresses
	for i := 1; i < 40; i++ {
		ip := fmt.Sprintf("10.0.%d.%d", rand.Intn(256), rand.Intn(256))
		port := rand.Intn(65535) + 1
		protocol := "tcp"
		if rand.Intn(2) == 1 {
			protocol = "udp"
		}
		bootstraper.AddAddress(ip, port, protocol)
	}

	err = bootstraper.AddPublicIPv4(80, "tcp")
	if err != nil {
		fmt.Println("Could not add public IPv4 address: ", err)
	}

	err = bootstraper.AddPublicIPv6(80, "tcp")
	if err != nil {
		fmt.Println("Could not add public IPv6 address: ", err)
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
