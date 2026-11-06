package diagnostics

import (
	"bufio"
	"container/heap"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

const pageSize = 200
const memoryLimit = 2000
const retention = 7 * 24 * time.Hour

type Event struct {
	ID              int64     `json:"id"`
	Time            time.Time `json:"time"`
	Level           string    `json:"level"`
	Message         string    `json:"msg"`
	Session         string    `json:"session"`
	Service         string    `json:"service"`
	Language        string    `json:"language"`
	Attempt         string    `json:"attempt"`
	Stage           string    `json:"stage"`
	ElapsedMS       int64     `json:"elapsed_ms"`
	Reason          string    `json:"reason"`
	HTTPStatus      int       `json:"http_status"`
	ProviderCode    int       `json:"provider_code"`
	ProviderSubcode int       `json:"provider_subcode"`
	Pages           int       `json:"pages"`
	Matches         int       `json:"matches"`
}

type Query struct {
	Search         string `json:"search"`
	Level          string `json:"level"`
	Service        string `json:"service"`
	Language       string `json:"language"`
	Attempt        string `json:"attempt"`
	CurrentSession bool   `json:"currentSession"`
	Before         int64  `json:"before"`
	After          int64  `json:"after"`
}

type Result struct {
	Events  []Event `json:"events"`
	More    bool    `json:"more"`
	Warning string  `json:"warning"`
}

type eventHeap []Event

func (h eventHeap) Len() int           { return len(h) }
func (h eventHeap) Less(i, j int) bool { return h[i].ID < h[j].ID }
func (h eventHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *eventHeap) Push(value any)    { *h = append(*h, value.(Event)) }
func (h *eventHeap) Pop() any {
	last := len(*h) - 1
	value := (*h)[last]
	*h = (*h)[:last]
	return value
}

type Store struct {
	mu                    sync.Mutex
	dir, session, warning string
	nextID                int64
	recent                []Event
	handler               slog.Handler
	writer                *lumberjack.Logger
	unlock                func()
	changed               chan struct{}
	done                  chan struct{}
	closed                bool
	listening             bool
}

// Open degrades to a bounded memory log when the local log directory is unavailable.
func Open(dir string) *Store {
	s := &Store{dir: dir, session: rand.Text(), changed: make(chan struct{}, 1), done: make(chan struct{})}
	s.handler = slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})
	if err := os.MkdirAll(dir, 0700); err != nil {
		s.warning = "Logs are available in memory only. The local log directory could not be opened."
		return s
	}
	unlock, err := lockDirectory(dir)
	if err != nil {
		s.warning = "Logs are available in memory only. Another Publisher instance may be using the log files."
		return s
	}
	s.unlock = unlock
	// Expire the active file by its last write time before rotating. Backup
	// timestamps are then managed by Lumberjack, including during quiet sessions.
	active := filepath.Join(dir, "publisher.jsonl")
	if info, err := os.Lstat(active); err == nil && info.Mode().IsRegular() && time.Since(info.ModTime()) > retention {
		if err := os.Remove(active); err != nil {
			s.warning = "An expired log file could not be removed."
		}
	}
	for _, file := range s.files() {
		scan(file, func(e Event) {
			if e.ID > s.nextID {
				s.nextID = e.ID
			}
		})
	}
	s.writer = &lumberjack.Logger{Filename: active, MaxSize: 10, MaxBackups: 5, MaxAge: 7}
	if err := s.writer.Rotate(); err != nil {
		s.warning = "Log rotation failed. Check access to the local log folder."
	}
	s.handler = slog.NewJSONHandler(s.writer, &slog.HandlerOptions{Level: slog.LevelDebug})
	go s.maintain()
	return s
}

func (s *Store) maintain() {
	timer := time.NewTicker(24 * time.Hour)
	defer timer.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-timer.C:
			s.mu.Lock()
			if !s.closed && s.writer.Rotate() != nil {
				s.warning = "Log rotation failed. Check access to the local log folder."
			}
			s.mu.Unlock()
		}
	}
}

func (s *Store) record(e Event, level slog.Level) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.nextID++
	e.ID, e.Session = s.nextID, s.session
	r := slog.NewRecord(e.Time, level, e.Message, 0)
	r.AddAttrs(
		slog.Int64("id", e.ID), slog.String("session", e.Session),
		slog.String("service", e.Service), slog.String("language", e.Language),
		slog.String("attempt", e.Attempt), slog.String("stage", e.Stage),
		slog.Int64("elapsed_ms", e.ElapsedMS), slog.String("reason", e.Reason),
		slog.Int("http_status", e.HTTPStatus), slog.Int("provider_code", e.ProviderCode),
		slog.Int("provider_subcode", e.ProviderSubcode), slog.Int("pages", e.Pages), slog.Int("matches", e.Matches))
	if s.handler.Handle(context.Background(), r) != nil {
		s.warning = "Some events could not be saved. Recent events remain available in memory."
	}
	if len(s.recent) == memoryLimit {
		copy(s.recent, s.recent[1:])
		s.recent = s.recent[:memoryLimit-1]
	}
	s.recent = append(s.recent, e)
	if s.listening {
		select {
		case s.changed <- struct{}{}:
		default:
		}
	}
}

func (s *Store) Listen(enabled bool)      { s.mu.Lock(); s.listening = enabled; s.mu.Unlock() }
func (s *Store) Changes() <-chan struct{} { return s.changed }
func (s *Store) Done() <-chan struct{}    { return s.done }
func (s *Store) Directory() string        { return s.dir }

func (s *Store) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.done)
	if s.writer != nil {
		_ = s.writer.Close()
	}
	if s.unlock != nil {
		s.unlock()
	}
}

func (s *Store) files() []string {
	items, _ := os.ReadDir(s.dir)
	type logFile struct {
		path string
		time time.Time
	}
	files := []logFile{}
	for _, item := range items {
		name := item.Name()
		if name != "publisher.jsonl" && !(strings.HasPrefix(name, "publisher-") && strings.HasSuffix(name, ".jsonl")) {
			continue
		}
		info, err := item.Info()
		if err != nil || !info.Mode().IsRegular() || info.Size() > 10*1024*1024 || time.Since(info.ModTime()) > retention {
			continue
		}
		files = append(files, logFile{filepath.Join(s.dir, name), info.ModTime()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].time.Before(files[j].time) })
	if len(files) > 6 {
		files = files[len(files)-6:]
	}
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.path)
	}
	return paths
}

func scan(path string, accept func(Event)) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	lines := bufio.NewScanner(io.LimitReader(file, 10*1024*1024))
	for lines.Scan() {
		var event Event
		if json.Unmarshal(lines.Bytes(), &event) == nil && validEvent(event) {
			accept(event)
		}
	}
}

// Validate messages against our vocabulary when reading disk. This also keeps
// arbitrary text inserted into a local file out of exports and the UI.
func validEvent(e Event) bool {
	message, known := messages[e.Stage]
	if !known || message != e.Message || len(e.Attempt) > 64 || len(e.Session) > 64 || e.ID < 1 || e.ID > 9007199254740990 || e.Reason != safeReason(e.Reason) {
		return false
	}
	for _, value := range []string{e.Attempt, e.Session, e.Reason} {
		for _, c := range value {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') {
				return false
			}
		}
	}
	switch e.Service {
	case "youtube", "tiktok", "instagram", "x":
	default:
		return false
	}
	return (e.Language == "es" || e.Language == "en") && (e.Level == "DEBUG" || e.Level == "INFO" || e.Level == "WARN" || e.Level == "ERROR")
}

func (s *Store) Query(q Query) (Result, error) {
	if len(q.Search) > 200 || len(q.Attempt) > 64 || q.Before < 0 || q.Before > 9007199254740991 || q.After < 0 || q.After > 9007199254740991 ||
		(q.Level != "" && q.Level != "DEBUG" && q.Level != "INFO" && q.Level != "WARN" && q.Level != "ERROR") ||
		(q.Service != "" && q.Service != "youtube" && q.Service != "tiktok" && q.Service != "instagram" && q.Service != "x") ||
		(q.Language != "" && q.Language != "es" && q.Language != "en") {
		return Result{}, errors.New("invalid diagnostic filter")
	}
	s.mu.Lock()
	recent := append([]Event(nil), s.recent...)
	maxID, session, warning := s.nextID, s.session, s.warning
	s.mu.Unlock()
	result := Result{Events: []Event{}, Warning: warning}
	matches := &eventHeap{}
	search := strings.ToLower(strings.TrimSpace(q.Search))
	seen := make(map[int64]bool, len(recent))
	for _, e := range recent {
		seen[e.ID] = true
	}
	cutoff := time.Now().Add(-retention)
	accept := func(e Event) {
		if e.ID > maxID || e.ID <= q.After || e.Time.Before(cutoff) || (q.Before > 0 && e.ID >= q.Before) {
			return
		}
		if q.CurrentSession && e.Session != session || q.Level != "" && e.Level != q.Level || q.Service != "" && e.Service != q.Service || q.Language != "" && e.Language != q.Language || q.Attempt != "" && e.Attempt != q.Attempt {
			return
		}
		if search != "" && !strings.Contains(strings.ToLower(e.Message+" "+e.Stage+" "+e.Reason+" "+e.Attempt+" "+e.Service+" "+e.Language+" "+strconv.Itoa(e.HTTPStatus)+" "+strconv.Itoa(e.ProviderCode)+" "+strconv.Itoa(e.ProviderSubcode)), search) {
			return
		}
		heap.Push(matches, e)
		if matches.Len() > pageSize {
			heap.Pop(matches)
			result.More = true
		}
	}
	if s.writer != nil {
		for _, file := range s.files() {
			scan(file, func(e Event) {
				if !seen[e.ID] {
					accept(e)
				}
			})
		}
	}
	for _, e := range recent {
		accept(e)
	}
	result.Events = []Event(*matches)
	sort.Slice(result.Events, func(i, j int) bool { return result.Events[i].ID > result.Events[j].ID })
	return result, nil
}
