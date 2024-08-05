package PathfinderBeacon

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/i5heu/PathfinderBeacon/pkg/auth"
	"github.com/i5heu/PathfinderBeacon/pkg/utils"
)

type PathfinderBeacon struct {
	nodes map[string][]string
	self  utils.RegisteringNode
	Key   *auth.Key
	mu    sync.RWMutex
}

func NewPathfinderBeacon(authConfig *auth.Key) (*PathfinderBeacon, error) {

	if authConfig.PrivateKey == nil {
		var err error
		authConfig, err = auth.GenerateKey()
		if err != nil {
			return nil, err
		}
	}

	instance := &PathfinderBeacon{
		Key:   authConfig,
		nodes: make(map[string][]string),
		self:  utils.RegisteringNode{},
	}

	instance.self.Room = instance.Key.GetRoomName()
	instance.self.PublicKey = instance.Key.PublicKeyToPemBase64()

	sign, err := instance.Key.GetRoomSignature()
	if err != nil {
		return nil, err
	}
	instance.self.RoomSignature = base64.StdEncoding.EncodeToString(sign)

	return instance, nil
}

func (p *PathfinderBeacon) AddAddress(ip string, port int, protocol string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.self.Addresses = append(p.self.Addresses, utils.RegisteringAddress{
		Protocol: protocol,
		Ip:       ip,
		Port:     port,
	})
}

func (p *PathfinderBeacon) GetAddresses() utils.RegisteringNode {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.self
}

func (p *PathfinderBeacon) PushAddresses() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	url := "https://pathfinderbeacon.net/register"

	jsonData, err := json.Marshal(p.self)
	if err != nil {
		return fmt.Errorf("Error marshalling JSON: %v\n", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Error creating request: %v\n", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v\n", err)
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Error reading response body: %v", err)
		}
		bodyString := string(bodyBytes)

		return fmt.Errorf("Error response: %v, %s", resp.Status, bodyString)
	}
	defer resp.Body.Close()

	return nil
}

func (p *PathfinderBeacon) PullNodes() error {
	roomName := p.Key.GetRoomName()
	roomNodeLinks, err := GetRoomNodeLinks(roomName)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, nodeLink := range roomNodeLinks {
		p.nodes[nodeLink], err = GetRoomNodeAddresses(nodeLink)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetRoomNodeLinks(roomName string) ([]string, error) {
	roomDomain := ".room.pathfinderbeacon.net"
	roomRequestDomain := roomName + roomDomain

	// make TXT DNS request to roomRequestDomain
	txtRecords, err := net.LookupTXT(roomRequestDomain)
	if err != nil {
		return nil, fmt.Errorf("Error looking up TXT records: %v\n", err)
	}

	return txtRecords, nil
}

func GetRoomNodeAddresses(nodeName string) ([]string, error) {
	nodeDomain := ".node.pathfinderbeacon.net"
	nodeRequestDomain := nodeName + nodeDomain

	// make TXT DNS request to roomRequestDomain
	txtRecords, err := net.LookupTXT(nodeRequestDomain)
	if err != nil {
		return nil, fmt.Errorf("Error looking up TXT records: %v\n", err)
	}

	return txtRecords, nil
}

func (p *PathfinderBeacon) GetNodes() map[string][]string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.nodes
}

func (p *PathfinderBeacon) GetRoomName() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.self.Room
}

func (p *PathfinderBeacon) GetRoomSignatureBase64() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.self.RoomSignature
}

func (p *PathfinderBeacon) GetPublicKeyBase64() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return base64.StdEncoding.EncodeToString([]byte(p.Key.PublicKeyToPem()))
}

func (p *PathfinderBeacon) AddPublicIPv4(port int, protocol string) error {
	resp, err := http.Get("https://ipv4.icanhazip.com")
	if err != nil {
		return fmt.Errorf("Error sending request: %v\n", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %v\n", err)
	}

	bodyString := strings.TrimSpace(string(bodyBytes))

	ip := net.ParseIP(bodyString)
	if ip == nil {
		return fmt.Errorf("Invalid IP address: %s", bodyString)
	}

	p.AddAddress(ip.String(), port, protocol)

	return nil
}

func (p *PathfinderBeacon) AddPublicIPv6(port int, protocol string) error {
	resp, err := http.Get("https://ipv6.icanhazip.com")
	if err != nil {
		return fmt.Errorf("Error sending request: %v\n", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %v\n", err)
	}

	bodyString := strings.TrimSpace(string(bodyBytes))

	ip := net.ParseIP(bodyString)
	if ip == nil {
		return fmt.Errorf("Invalid IP address: %s", bodyString)
	}

	p.AddAddress(ip.String(), port, protocol)

	return nil
}
