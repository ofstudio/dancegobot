package teletest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tidwall/gjson"
	tele "gopkg.in/telebot.v4"
)

const defaultTimeout = 2 * time.Second

// Server is a fake Telegram Bot API server for tests.
type Server struct {
	t testing.TB

	srv     *httptest.Server
	botUser *tele.User
	timeout time.Duration

	mu        sync.Mutex
	requests  []Request
	consumed  []bool
	waiters   []*waiter
	responses map[string][]Response
	ignored   map[string]bool
	updateID  int
	messageID int
	updates   chan tele.Update
	closed    bool
	closeOnce sync.Once
	nextMsgID int
}

// Option configures a Server.
type Option func(*Server)

// WithBotUser sets the user returned from getMe.
func WithBotUser(user *tele.User) Option {
	return func(s *Server) {
		if user != nil {
			u := *user
			s.botUser = &u
		}
	}
}

// WithTimeout sets the default timeout used by Wait.
func WithTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		if timeout > 0 {
			s.timeout = timeout
		}
	}
}

// New starts a fake Telegram Bot API server.
func New(t testing.TB, opts ...Option) *Server {
	t.Helper()
	s := &Server{
		t:       t,
		timeout: defaultTimeout,
		botUser: &tele.User{
			ID:        1234567890,
			IsBot:     true,
			FirstName: "test_bot",
			Username:  "test_bot",
		},
		responses: make(map[string][]Response),
		ignored:   make(map[string]bool),
		updates:   make(chan tele.Update, 100),
		nextMsgID: 1,
	}
	for _, opt := range opts {
		opt(s)
	}
	s.srv = httptest.NewServer(http.HandlerFunc(s.handle))
	t.Cleanup(s.Close)
	return s
}

// URL returns the base Bot API URL to pass to telebot.Settings.URL.
func (s *Server) URL() string {
	return s.srv.URL
}

// Close stops the fake server.
func (s *Server) Close() {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.closed = true
		close(s.updates)
		s.mu.Unlock()
		s.srv.Close()
	})
}

// Ignore excludes methods from AssertNoUnexpected.
func (s *Server) Ignore(methods ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, method := range methods {
		s.ignored[method] = true
	}
}

// Respond queues a successful Bot API response for method.
func (s *Server) Respond(method string, result any) {
	s.queueResponse(method, Response{OK: true, Result: result})
}

// RespondError queues an error Bot API response for method.
func (s *Server) RespondError(method string, code int, description string) {
	s.queueResponse(method, Response{OK: false, ErrorCode: code, Description: description})
}

func (s *Server) queueResponse(method string, resp Response) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.responses[method] = append(s.responses[method], resp)
}

// Push queues an update for the fake getUpdates endpoint.
func (s *Server) Push(update tele.Update) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.updates <- update
}

// Message wraps a Telegram message into an update and fills missing IDs.
func (s *Server) Message(msg *tele.Message) tele.Update {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateID++
	if msg != nil && msg.ID == 0 {
		s.messageID++
		msg.ID = s.messageID
	}
	return tele.Update{ID: s.updateID, Message: msg}
}

// InlineQuery wraps an inline query into an update and fills missing IDs.
func (s *Server) InlineQuery(query *tele.Query) tele.Update {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateID++
	if query != nil && query.ID == "" {
		query.ID = fmt.Sprintf("inline_query_id_%d", s.updateID)
	}
	return tele.Update{ID: s.updateID, Query: query}
}

// InlineResult wraps a chosen inline result into an update and fills missing IDs.
func (s *Server) InlineResult(result *tele.InlineResult) tele.Update {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateID++
	if result != nil && result.MessageID == "" {
		result.MessageID = fmt.Sprintf("inline_message_id_%d", s.updateID)
	}
	return tele.Update{ID: s.updateID, InlineResult: result}
}

// CallbackQuery wraps a callback query into an update and fills missing IDs.
func (s *Server) CallbackQuery(callback *tele.Callback) tele.Update {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateID++
	if callback != nil && callback.ID == "" {
		callback.ID = fmt.Sprintf("callback_query_id_%d", s.updateID)
	}
	return tele.Update{ID: s.updateID, Callback: callback}
}

// Wait waits for a request with method.
func (s *Server) Wait(method string, timeout ...time.Duration) Request {
	return s.WaitFor(method, nil, timeout...)
}

// WaitFor waits for a request with method that satisfies check.
func (s *Server) WaitFor(method string, check func(Request) bool, timeout ...time.Duration) Request {
	s.t.Helper()
	waitTimeout := s.timeout
	if len(timeout) > 0 && timeout[0] > 0 {
		waitTimeout = timeout[0]
	}
	ch := make(chan Request, 1)

	s.mu.Lock()
	if req, ok := s.consumeLocked(method, check); ok {
		s.mu.Unlock()
		return req
	}
	w := &waiter{method: method, check: check, ch: ch}
	s.waiters = append(s.waiters, w)
	s.mu.Unlock()

	select {
	case req := <-ch:
		return req
	case <-time.After(waitTimeout):
		s.t.Fatalf("timed out waiting for Telegram API request %q", method)
		return Request{}
	}
}

// AssertNoUnexpected fails if there are unconsumed, non-ignored requests.
func (s *Server) AssertNoUnexpected() {
	s.t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	var methods []string
	for i, req := range s.requests {
		if s.consumed[i] || s.ignored[req.Method] {
			continue
		}
		methods = append(methods, req.Method+" "+req.PrettyBody())
	}
	if len(methods) > 0 {
		s.t.Fatalf("unexpected Telegram API requests: %s", strings.Join(methods, ", "))
	}
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	method := path.Base(r.URL.Path)
	if method == "." || method == "/" {
		http.NotFound(w, r)
		return
	}
	if method == "getUpdates" {
		s.handleGetUpdates(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIResponse(w, Response{OK: false, ErrorCode: 500, Description: err.Error()})
		return
	}
	_ = r.Body.Close()
	req := Request{Method: method, Body: body}
	req.JSON = gjson.ParseBytes(body)
	s.record(req)

	resp := s.nextResponse(method)
	writeAPIResponse(w, resp)
}

func (s *Server) handleGetUpdates(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()
	req := Request{Method: "getUpdates", Body: body, JSON: gjson.ParseBytes(body)}
	s.record(req)

	select {
	case update, ok := <-s.updates:
		if !ok {
			writeAPIResponse(w, Response{OK: true, Result: []tele.Update{}})
			return
		}
		writeAPIResponse(w, Response{OK: true, Result: []tele.Update{update}})
	case <-r.Context().Done():
		writeAPIResponse(w, Response{OK: true, Result: []tele.Update{}})
	case <-time.After(10 * time.Millisecond):
		writeAPIResponse(w, Response{OK: true, Result: []tele.Update{}})
	}
}

func (s *Server) record(req Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests = append(s.requests, req)
	s.consumed = append(s.consumed, false)
	s.satisfyWaitersLocked()
}

func (s *Server) satisfyWaitersLocked() {
	for len(s.waiters) > 0 {
		matched := false
		for wi, w := range s.waiters {
			if req, ok := s.consumeLocked(w.method, w.check); ok {
				s.waiters = append(s.waiters[:wi], s.waiters[wi+1:]...)
				w.ch <- req
				matched = true
				break
			}
		}
		if !matched {
			return
		}
	}
}

func (s *Server) consumeLocked(method string, check func(Request) bool) (Request, bool) {
	for i, req := range s.requests {
		if s.consumed[i] || req.Method != method {
			continue
		}
		if check != nil && !check(req) {
			continue
		}
		s.consumed[i] = true
		return req, true
	}
	return Request{}, false
}

func (s *Server) nextResponse(method string) Response {
	s.mu.Lock()
	defer s.mu.Unlock()
	if queued := s.responses[method]; len(queued) > 0 {
		resp := queued[0]
		s.responses[method] = queued[1:]
		return resp
	}
	return s.defaultResponse(method)
}

func (s *Server) defaultResponse(method string) Response {
	switch method {
	case "getMe":
		return Response{OK: true, Result: s.botUser}
	case "sendMessage":
		id := s.nextMsgID
		s.nextMsgID++
		return Response{OK: true, Result: tele.Message{ID: id}}
	case "editMessageText", "editMessageReplyMarkup":
		return Response{OK: true, Result: true}
	case "answerInlineQuery", "answerCallbackQuery", "setMyCommands", "deleteWebhook":
		return Response{OK: true, Result: true}
	default:
		return Response{OK: true, Result: true}
	}
}

func writeAPIResponse(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

type waiter struct {
	method string
	check  func(Request) bool
	ch     chan Request
}

// Response is a Telegram Bot API response.
type Response struct {
	OK          bool   `json:"ok"`
	Result      any    `json:"result,omitempty"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}

// Request is a recorded Telegram Bot API request.
type Request struct {
	Method string
	Body   []byte
	JSON   gjson.Result
}

// String returns a JSON field as a string.
func (r Request) String(path string) string {
	return r.JSON.Get(path).String()
}

// Int returns a JSON field as an int.
func (r Request) Int(path string) int {
	return int(r.JSON.Get(path).Int())
}

// Bool returns a JSON field as a bool.
func (r Request) Bool(path string) bool {
	return r.JSON.Get(path).Bool()
}

// ReplyMarkup parses the reply_markup field.
func (r Request) ReplyMarkup() gjson.Result {
	rm := r.JSON.Get("reply_markup")
	if rm.Type == gjson.String {
		return gjson.Parse(rm.String())
	}
	return rm
}

// InlineKeyboardRaw returns the raw inline keyboard JSON.
func (r Request) InlineKeyboardRaw() string {
	return r.ReplyMarkup().Get("inline_keyboard").Raw
}

// PrettyBody returns the request body in a stable, readable form.
func (r Request) PrettyBody() string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, r.Body, "", "  "); err != nil {
		return string(r.Body)
	}
	return buf.String()
}

// ChatID returns chat_id as string because Telegram accepts both numeric and
// string chat identifiers.
func (r Request) ChatID() string {
	return r.String("chat_id")
}

// ChatIDInt returns chat_id as int64.
func (r Request) ChatIDInt() int64 {
	chatID := r.JSON.Get("chat_id")
	if chatID.Type == gjson.Number {
		return chatID.Int()
	}
	id, _ := strconv.ParseInt(chatID.String(), 10, 64)
	return id
}
