//session_memory.go

package Services

import (
    "sync"
    "time"
)

type ChatMessage struct {
    Role    string // "user" | "assistant"
    Content string
}

type sessionState struct {
    History  []ChatMessage
    LastSeen time.Time
}

type SessionMemory struct {
    mu       sync.RWMutex
    ttl      time.Duration
    maxTurns int
    sessions map[string]*sessionState
}

func NewSessionMemory(ttl time.Duration, maxTurns int) *SessionMemory {
    sm := &SessionMemory{
        ttl:      ttl,
        maxTurns: maxTurns,
        sessions: make(map[string]*sessionState),
    }
    go sm.gcLoop()
    return sm
}

func (sm *SessionMemory) gcLoop() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        sm.GC()
    }
}

func (sm *SessionMemory) GC() {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    now := time.Now()
    for k, v := range sm.sessions {
        if now.Sub(v.LastSeen) > sm.ttl {
            delete(sm.sessions, k)
        }
    }
}

func (sm *SessionMemory) Get(sid string) []ChatMessage {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    if st, ok := sm.sessions[sid]; ok {
        out := make([]ChatMessage, len(st.History))
        copy(out, st.History)
        return out
    }
    return nil
}

func (sm *SessionMemory) Append(sid string, msg ChatMessage) {
    if sid == "" {
        return
    }
    sm.mu.Lock()
    defer sm.mu.Unlock()

    st, ok := sm.sessions[sid]
    if !ok {
        st = &sessionState{LastSeen: time.Now()}
        sm.sessions[sid] = st
    }
    st.LastSeen = time.Now()
    st.History = append(st.History, msg)

    maxMsgs := sm.maxTurns * 2
    if maxMsgs > 0 && len(st.History) > maxMsgs {
        st.History = st.History[len(st.History)-maxMsgs:]
    }
}

func (sm *SessionMemory) Reset(sid string) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    delete(sm.sessions, sid)
}
