package sync

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/yoshapihoff/smart-clipboard/internal/types"
)

type SyncData struct {
	History []types.ClipboardItem `json:"history"`
}

type SyncManager struct {
	conn         *net.UDPConn
	serverAddr   *net.UDPAddr
	sendEnabled  bool
	receiveEnabled bool
	mu           sync.Mutex
	historyChan  chan<- []types.ClipboardItem
}

func NewSyncManager(listenPort int, sendToAddr string, historyChan chan<- []types.ClipboardItem) (*SyncManager, error) {
	sm := &SyncManager{
		sendEnabled:  sendToAddr != "",
		receiveEnabled: listenPort > 0,
		historyChan:  historyChan,
	}

	var err error

	if sm.receiveEnabled {
		listenAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", listenPort))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve listen address: %w", err)
		}

		sm.conn, err = net.ListenUDP("udp", listenAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to listen on UDP: %w", err)
		}
	}

	if sm.sendEnabled && sendToAddr != "" {
		sm.serverAddr, err = net.ResolveUDPAddr("udp", sendToAddr)
		if err != nil {
			if sm.conn != nil {
				sm.conn.Close()
			}
			return nil, fmt.Errorf("failed to resolve server address: %w", err)
		}
	}

	return sm, nil
}

func (sm *SyncManager) Start() {
	if sm.receiveEnabled {
		go sm.receiveLoop()
	}
}

func (sm *SyncManager) Stop() {
	if sm.conn != nil {
		sm.conn.Close()
	}
}

func (sm *SyncManager) SendHistory(history []types.ClipboardItem) error {
	if !sm.sendEnabled || sm.serverAddr == nil {
		return nil
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	data := SyncData{
		History: history,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, sm.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to dial UDP: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	return nil
}

func (sm *SyncManager) receiveLoop() {
	if !sm.receiveEnabled || sm.conn == nil {
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

	log.Printf("Received %d history items from %s", len(syncData.History), addr)

	if sm.historyChan != nil {
		sm.historyChan <- syncData.History
	}
}

func (sm *SyncManager) SetSendEnabled(enabled bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sendEnabled = enabled
}

func (sm *SyncManager) SetReceiveEnabled(enabled bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.receiveEnabled = enabled
}
