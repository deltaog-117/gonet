package appstate

import "sync"

// AppState holds the global application state.
type AppState struct {
	mu          sync.RWMutex
	currentSSID string
	connected   bool
}

// New creates a new AppState.
func New() *AppState {
	return &AppState{}
}

// SetConnected updates the connection state.
func (s *AppState) SetConnected(ssid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentSSID = ssid
	s.connected = true
}

// SetDisconnected updates the disconnection state.
func (s *AppState) SetDisconnected() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentSSID = ""
	s.connected = false
}

// IsConnected returns the current connection status.
func (s *AppState) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

// GetCurrentSSID returns the current SSID.
func (s *AppState) GetCurrentSSID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentSSID
}
