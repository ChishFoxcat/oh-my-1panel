package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/ChishFoxcat/oh-my-1panel/backend/utils/panel"
)

const cleanInterval = time.Minute

// Session 是浏览器侧会话，绑定了一个持有面板凭据的面板客户端。
type Session struct {
	ID           string
	CSRFToken    string
	Name         string
	Role         string
	Panel        *panel.Client
	LoginAt      time.Time
	LastActiveAt time.Time
	ExpiresAt    time.Time
}

// IsAdmin 判断当前会话是否为面板管理员。
func (s *Session) IsAdmin() bool {
	return s.Role == "ADMIN"
}

// Pending 是等待动态验证码的中间会话，验证通过后才会生成正式会话。
type Pending struct {
	ID         string
	Name       string
	MFASession string
	Panel      *panel.Client
	ExpiresAt  time.Time
}

// Store 是内存会话存储，带滑动过期与后台清理。
type Store struct {
	mu         sync.RWMutex
	sessions   map[string]*Session
	pending    map[string]*Pending
	ttl        time.Duration
	pendingTTL time.Duration
	stop       chan struct{}
}

// NewStore 创建会话存储并启动清理协程。
func NewStore(ttl, pendingTTL time.Duration) *Store {
	store := &Store{
		sessions:   make(map[string]*Session),
		pending:    make(map[string]*Pending),
		ttl:        ttl,
		pendingTTL: pendingTTL,
		stop:       make(chan struct{}),
	}
	go store.cleanLoop()
	return store
}

// Close 停止清理协程。
func (s *Store) Close() {
	close(s.stop)
}

// Create 生成正式会话，同时生成 CSRF 令牌。
func (s *Store) Create(client *panel.Client, name, role string) (*Session, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}
	csrfToken, err := newID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	item := &Session{
		ID:           id,
		CSRFToken:    csrfToken,
		Name:         name,
		Role:         role,
		Panel:        client,
		LoginAt:      now,
		LastActiveAt: now,
		ExpiresAt:    now.Add(s.ttl),
	}
	s.mu.Lock()
	s.sessions[id] = item
	s.mu.Unlock()
	return item, nil
}

// Get 读取会话并按需滑动续期，过期或不存在时返回 false。
func (s *Store) Get(id string) (*Session, bool) {
	if id == "" {
		return nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	now := time.Now()
	if now.After(item.ExpiresAt) {
		delete(s.sessions, id)
		item.Panel.Close()
		return nil, false
	}
	item.LastActiveAt = now
	item.ExpiresAt = now.Add(s.ttl)
	return item, true
}

// Drop 删除会话并释放面板连接。
func (s *Store) Drop(id string) {
	s.mu.Lock()
	item, ok := s.sessions[id]
	delete(s.sessions, id)
	s.mu.Unlock()
	if ok {
		item.Panel.Close()
	}
}

// CreatePending 登记等待动态验证码的中间会话。
func (s *Store) CreatePending(client *panel.Client, name, mfaSession string) (*Pending, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}
	item := &Pending{
		ID:         id,
		Name:       name,
		MFASession: mfaSession,
		Panel:      client,
		ExpiresAt:  time.Now().Add(s.pendingTTL),
	}
	s.mu.Lock()
	s.pending[id] = item
	s.mu.Unlock()
	return item, nil
}

// GetPending 读取中间会话（不消费），过期或不存在时返回 false。
// 动态验证码输入错误时面板会保留会话，这里同样保留以便用户重试。
func (s *Store) GetPending(id string) (*Pending, bool) {
	if id == "" {
		return nil, false
	}
	s.mu.Lock()
	item, ok := s.pending[id]
	if ok && time.Now().After(item.ExpiresAt) {
		delete(s.pending, id)
		ok = false
	}
	s.mu.Unlock()
	if !ok {
		return nil, false
	}
	return item, true
}

// DropPending 删除中间会话并释放面板连接。
func (s *Store) DropPending(id string) {
	s.mu.Lock()
	item, ok := s.pending[id]
	delete(s.pending, id)
	s.mu.Unlock()
	if ok {
		item.Panel.Close()
	}
}

func (s *Store) cleanLoop() {
	ticker := time.NewTicker(cleanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.clean()
		}
	}
}

func (s *Store) clean() {
	now := time.Now()
	s.mu.Lock()
	for id, item := range s.sessions {
		if now.After(item.ExpiresAt) {
			delete(s.sessions, id)
			item.Panel.Close()
		}
	}
	for id, item := range s.pending {
		if now.After(item.ExpiresAt) {
			delete(s.pending, id)
			item.Panel.Close()
		}
	}
	s.mu.Unlock()
}

func newID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("生成会话标识失败：%w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
