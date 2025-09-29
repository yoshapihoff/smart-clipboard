package sync

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yoshapihoff/smart-clipboard/internal/constants"
	"github.com/yoshapihoff/smart-clipboard/internal/types"
)

type SyncData struct {
	History []types.ClipboardItem `json:"history"`
	Type    string                `json:"type"` // "history" or "request_history"
}

type SyncManager struct {
	conn          *net.UDPConn
	broadcastConn *net.UDPConn
	serverAddrs   []*net.UDPAddr
	mu            sync.Mutex
	historyChan   chan<- []types.ClipboardItem
	stopChan      chan struct{}

	// Callback for getting current history
	getHistoryFunc func() []types.ClipboardItem
}

func NewSyncManager(historyChan chan<- []types.ClipboardItem) (*SyncManager, error) {
	sm := &SyncManager{
		historyChan: historyChan,
		stopChan:    make(chan struct{}),
	}

	// Setup main sync listener
	listenAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", constants.SyncPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve listen address: %w", err)
	}

	sm.conn, err = net.ListenUDP("udp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UDP: %w", err)
	}

	// Setup broadcast listener for discovery
	broadcastAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", constants.SyncBroadcastPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve broadcast address: %w", err)
	}

	sm.broadcastConn, err = net.ListenUDP("udp", broadcastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on broadcast UDP: %w", err)
	}

	// Start server discovery
	go sm.discoverServers()

	return sm, nil
}

func (sm *SyncManager) discoverServers() {
	// Start discovery loops
	go sm.broadcastDiscovery()
	go sm.listenForDiscovery()

	go sm.receiveLoop()
}

func (sm *SyncManager) Stop() {
	close(sm.stopChan)
	if sm.conn != nil {
		sm.conn.Close()
	}
	if sm.broadcastConn != nil {
		sm.broadcastConn.Close()
	}
}

// SetHistoryCallback sets the callback function to get current history
func (sm *SyncManager) SetHistoryCallback(callback func() []types.ClipboardItem) {
	sm.getHistoryFunc = callback
}

func (sm *SyncManager) SendHistory(history []types.ClipboardItem) error {
	sm.mu.Lock()
	serverCount := len(sm.serverAddrs)
	sm.mu.Unlock()
	
	log.Printf("SendHistory called with %d items, sending to %d servers", len(history), serverCount)
	if serverCount == 0 {
		log.Printf("No servers discovered, skipping history send")
		return nil
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	data := SyncData{
		History: history,
		Type:    "history",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// Send to all discovered servers
	for _, serverAddr := range sm.serverAddrs {
		conn, err := net.DialUDP("udp", nil, serverAddr)
		if err != nil {
			log.Printf("Failed to dial UDP to %s: %v", serverAddr, err)
			continue
		}

		_, err = conn.Write(jsonData)
		conn.Close()
		if err != nil {
			log.Printf("Failed to send data to %s: %v", serverAddr, err)
			continue
		}

		log.Printf("Sent %d history items to %s", len(history), serverAddr)
	}

	return nil
}

func (sm *SyncManager) receiveLoop() {
	if sm.conn == nil {
		return
	}

	buffer := make([]byte, 65535) // 64KB max UDP packet size

	for {
		n, addr, err := sm.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		go sm.processReceivedData(buffer[:n], addr)
	}
}

func (sm *SyncManager) processReceivedData(data []byte, addr *net.UDPAddr) {
	var syncData SyncData

	// Try JSON first
	err := json.Unmarshal(data, &syncData)
	if err != nil {
		// Fallback to gob
		decoder := gob.NewDecoder(bytes.NewReader(data))
		err = decoder.Decode(&syncData)
		if err != nil {
			log.Printf("Failed to unmarshal data from %s: %v", addr, err)
			return
		}
	}

	switch syncData.Type {
	case "history":
		log.Printf("Received %d history items from %s", len(syncData.History), addr)
		if sm.historyChan != nil {
			sm.historyChan <- syncData.History
		}
	default:
		log.Printf("Received unknown sync data type from %s: %s", addr, syncData.Type)
	}
}

func (sm *SyncManager) broadcastDiscovery() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.sendDiscoveryBroadcast()
		case <-sm.stopChan:
			return
		}
	}
}

func (sm *SyncManager) sendDiscoveryBroadcast() {
	// Get local network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Failed to get network interfaces: %v", err)
		return
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // Skip IPv6
			}

			// Create broadcast address
			broadcastAddr := &net.UDPAddr{
				IP:   net.IPv4bcast,
				Port: constants.SyncBroadcastPort,
			}

			// Send discovery message
			message := fmt.Sprintf("%s:%d", constants.DiscoveryMagicHeader, constants.SyncPort)
			conn, err := net.DialUDP("udp", nil, broadcastAddr)
			if err != nil {
				continue
			}

			_, err = conn.Write([]byte(message))
			conn.Close()
			if err != nil {
				continue
			}

			log.Printf("Sent discovery broadcast on %s", iface.Name)
		}
	}
}

func (sm *SyncManager) listenForDiscovery() {
	if sm.broadcastConn == nil {
		return
	}

	buffer := make([]byte, 1024)

	for {
		select {
		case <-sm.stopChan:
			return
		default:
			n, addr, err := sm.broadcastConn.ReadFromUDP(buffer)
			if err != nil {
				continue
			}

			message := string(buffer[:n])
			log.Printf("Received discovery message: %s from %s", message, addr)
			if strings.HasPrefix(message, constants.DiscoveryMagicHeader) {
				// Extract port from message
				parts := strings.Split(message, ":")
				if len(parts) != 2 {
					continue
				}

				port, err := strconv.Atoi(parts[1])
				if err != nil {
					continue
				}

				// Create server address
				serverAddr := &net.UDPAddr{
					IP:   addr.IP,
					Port: port,
				}

				// Add to server list if not already present
				sm.mu.Lock()
				found := false
				for _, existing := range sm.serverAddrs {
					if existing.IP.Equal(serverAddr.IP) && existing.Port == serverAddr.Port {
						found = true
						break
					}
				}

				if !found {
					sm.serverAddrs = append(sm.serverAddrs, serverAddr)
					log.Printf("Discovered server at %s:%d, total servers: %d", serverAddr.IP, serverAddr.Port, len(sm.serverAddrs))

					// Immediately send our history to the discovered server
					history := sm.getHistoryFunc()
					go sm.sendHistoryToServer(serverAddr, history)
				} else {
					log.Printf("Server %s:%d already known", serverAddr.IP, serverAddr.Port)
				}
				sm.mu.Unlock()
			}
		}
	}
}

// sendHistoryToServer sends current history to a specific server
func (sm *SyncManager) sendHistoryToServer(serverAddr *net.UDPAddr, history []types.ClipboardItem) {
	data := SyncData{
		History: history,
		Type:    "history",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal history for server %s: %v", serverAddr, err)
		return
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Printf("Failed to dial UDP to %s: %v", serverAddr, err)
		return
	}
	defer conn.Close()

	_, err = conn.Write(jsonData)
	if err != nil {
		log.Printf("Failed to send history to %s: %v", serverAddr, err)
		return
	}

	log.Printf("Sent %d history items to %s", len(history), serverAddr)
}
