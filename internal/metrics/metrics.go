package metrics

import (
	"sync"
	"time"
)

// Metrics holds application metrics
type Metrics struct {
	mu                    sync.RWMutex
	startTime            time.Time
	commandsExecuted     int64
	messagesProcessed    int64
	gamesScraped         int64
	errors               int64
	serversJoined        int64
	serversLeft          int64
	lastScrapeTime       time.Time
	lastScrapeSuccess    bool
	lastScrapeDuration   time.Duration
	activeConnections    int64
	totalMemoryUsage     int64
}

// New creates a new metrics instance
func New() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

// GetUptime returns the application uptime
func (m *Metrics) GetUptime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return time.Since(m.startTime)
}

// IncrementCommandsExecuted increments the commands executed counter
func (m *Metrics) IncrementCommandsExecuted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commandsExecuted++
}

// GetCommandsExecuted returns the number of commands executed
func (m *Metrics) GetCommandsExecuted() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.commandsExecuted
}

// IncrementMessagesProcessed increments the messages processed counter
func (m *Metrics) IncrementMessagesProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messagesProcessed++
}

// GetMessagesProcessed returns the number of messages processed
func (m *Metrics) GetMessagesProcessed() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.messagesProcessed
}

// IncrementGamesScraped increments the games scraped counter
func (m *Metrics) IncrementGamesScraped(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gamesScraped += count
}

// GetGamesScraped returns the number of games scraped
func (m *Metrics) GetGamesScraped() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.gamesScraped
}

// IncrementErrors increments the error counter
func (m *Metrics) IncrementErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors++
}

// GetErrors returns the number of errors
func (m *Metrics) GetErrors() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errors
}

// IncrementServersJoined increments the servers joined counter
func (m *Metrics) IncrementServersJoined() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.serversJoined++
}

// GetServersJoined returns the number of servers joined
func (m *Metrics) GetServersJoined() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.serversJoined
}

// IncrementServersLeft increments the servers left counter
func (m *Metrics) IncrementServersLeft() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.serversLeft++
}

// GetServersLeft returns the number of servers left
func (m *Metrics) GetServersLeft() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.serversLeft
}

// SetLastScrapeTime sets the last scrape time and success status
func (m *Metrics) SetLastScrapeTime(success bool, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastScrapeTime = time.Now()
	m.lastScrapeSuccess = success
	m.lastScrapeDuration = duration
}

// GetLastScrapeInfo returns the last scrape information
func (m *Metrics) GetLastScrapeInfo() (time.Time, bool, time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastScrapeTime, m.lastScrapeSuccess, m.lastScrapeDuration
}

// SetActiveConnections sets the number of active connections
func (m *Metrics) SetActiveConnections(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeConnections = count
}

// GetActiveConnections returns the number of active connections
func (m *Metrics) GetActiveConnections() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeConnections
}

// SetMemoryUsage sets the memory usage
func (m *Metrics) SetMemoryUsage(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalMemoryUsage = bytes
}

// GetMemoryUsage returns the memory usage
func (m *Metrics) GetMemoryUsage() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalMemoryUsage
}

// Summary returns a summary of all metrics
func (m *Metrics) Summary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return map[string]interface{}{
		"uptime":              m.GetUptime().String(),
		"commands_executed":   m.commandsExecuted,
		"messages_processed":  m.messagesProcessed,
		"games_scraped":       m.gamesScraped,
		"errors":              m.errors,
		"servers_joined":      m.serversJoined,
		"servers_left":        m.serversLeft,
		"last_scrape_time":    m.lastScrapeTime,
		"last_scrape_success": m.lastScrapeSuccess,
		"last_scrape_duration": m.lastScrapeDuration.String(),
		"active_connections":  m.activeConnections,
		"memory_usage_bytes":  m.totalMemoryUsage,
	}
}

// Global metrics instance
var globalMetrics = New()

// Package-level functions for convenience
func IncrementCommandsExecuted() {
	globalMetrics.IncrementCommandsExecuted()
}

func IncrementMessagesProcessed() {
	globalMetrics.IncrementMessagesProcessed()
}

func IncrementGamesScraped(count int64) {
	globalMetrics.IncrementGamesScraped(count)
}

func IncrementErrors() {
	globalMetrics.IncrementErrors()
}

func IncrementServersJoined() {
	globalMetrics.IncrementServersJoined()
}

func IncrementServersLeft() {
	globalMetrics.IncrementServersLeft()
}

func SetLastScrapeTime(success bool, duration time.Duration) {
	globalMetrics.SetLastScrapeTime(success, duration)
}

func GetMetrics() *Metrics {
	return globalMetrics
}