// Package pjsip provides a loopback SIP MESSAGE endpoint backed by PJPROJECT.
// Native calls run on a dedicated OS thread. Only one endpoint may be active
// in a process; use separate processes for separate SIP identities.
package pjsip

/*
#cgo windows LDFLAGS: -L${SRCDIR}/../build/lib -lpjgo
#include <stdlib.h>
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
	"unsafe"
)

var active sync.Mutex
var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)
var ErrClosed = errors.New("SIP endpoint closed")

type Config struct {
	Port     int
	Username string
}

// Event is either a received message or a final SIP delivery result.
// A 200 status indicates acceptance by the remote endpoint, not that it was read.
type Event struct {
	Kind        string `json:"kind"`
	ID          uint32 `json:"id,omitempty"`
	Peer        string `json:"peer"`
	ReplyTo     string `json:"reply_to,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Body        string `json:"body"`
	Code        int    `json:"code,omitempty"`
}

type request struct {
	uri, body string
	result    chan sendResult
}
type sendResult struct {
	id  uint32
	err error
}

type Endpoint struct {
	requests chan request
	events   chan Event
	stop     chan struct{}
	done     chan struct{}
	once     sync.Once
	err      error // written before done closes
}

func Version() string { return C.GoString(C.pjgo_version()) }

func nativeError(code C.int) error {
	if code == 0 {
		return nil
	}
	var text [256]C.char
	C.pjgo_error(code, &text[0], C.uint(len(text)))
	return fmt.Errorf("PJSIP %d: %s", int(code), C.GoString(&text[0]))
}

func New(config Config) (*Endpoint, error) {
	if config.Port < 1 || config.Port > 65535 {
		return nil, errors.New("port must be 1..65535")
	}
	if !usernamePattern.MatchString(config.Username) {
		return nil, errors.New("username must be 1..64 ASCII letters, digits, _, . or -")
	}
	if !active.TryLock() {
		return nil, errors.New("only one SIP endpoint per process is supported")
	}
	ep := &Endpoint{requests: make(chan request), events: make(chan Event, 128), stop: make(chan struct{}), done: make(chan struct{})}
	ready := make(chan error, 1)
	go ep.run(config, ready)
	if err := <-ready; err != nil {
		<-ep.done
		return nil, err
	}
	return ep, nil
}

func (ep *Endpoint) run(config Config, ready chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer func() {
		close(ep.events)
		active.Unlock()
		close(ep.done)
	}()
	username := C.CString(config.Username)
	err := nativeError(C.pjgo_start(C.uint(config.Port), username))
	C.free(unsafe.Pointer(username))
	ready <- err
	if err != nil {
		ep.err = err
		return
	}
	defer C.pjgo_stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	var id uint32
	var event C.pjgo_event
	for {
		select {
		case <-ep.stop:
			return
		case req := <-ep.requests:
			id++
			if id == 0 {
				id++
			}
			uri := C.CString(req.uri)
			body := C.CString(req.body)
			err := nativeError(C.pjgo_send(uri, body, C.uint(len(req.body)), C.uint(id)))
			C.free(unsafe.Pointer(uri))
			C.free(unsafe.Pointer(body))
			req.result <- sendResult{id, err}
		case <-ticker.C:
			if result := C.pjgo_poll(0); result < 0 {
				ep.err = errors.New("native event polling failed or event queue overflowed")
				return
			}
		}
		for C.pjgo_pop(&event) != 0 {
			kind := "message"
			if event.kind == 2 {
				kind = "status"
			}
			value := Event{Kind: kind, ID: uint32(event.id), Code: int(event.code), Peer: C.GoString(&event.peer[0]), ReplyTo: C.GoString(&event.reply_to[0]), ContentType: C.GoString(&event.content_type[0]), Body: C.GoStringN(&event.body[0], C.int(event.body_len))}
			select {
			case ep.events <- value:
			default:
				ep.err = errors.New("event consumer too slow: event channel overflow")
				return
			}
		}
	}
}

// Send submits a UTF-8 text message; observe Events for its delivery result.
// The initial binding limits bodies to 2048 bytes for local UDP messaging.
func (ep *Endpoint) Send(uri, body string) (uint32, error) {
	if len(uri) > 2000 || strings.ContainsAny(uri, "\x00\r\n") || !strings.HasPrefix(uri, "sip:") {
		return 0, errors.New("invalid SIP destination URI")
	}
	if len(body) > 2048 || !utf8.ValidString(body) {
		return 0, errors.New("message must be valid UTF-8, at most 2048 bytes")
	}
	req := request{uri: uri, body: body, result: make(chan sendResult, 1)}
	select {
	case <-ep.done:
		return 0, ErrClosed
	case <-ep.stop:
		return 0, ErrClosed
	case ep.requests <- req:
	}
	result := <-req.result
	return result.id, result.err
}

func (ep *Endpoint) Events() <-chan Event  { return ep.events }
func (ep *Endpoint) Done() <-chan struct{} { return ep.done }

// Err reports a terminal failure after Done closes, or nil while running.
func (ep *Endpoint) Err() error {
	select {
	case <-ep.done:
		return ep.err
	default:
		return nil
	}
}
func (ep *Endpoint) Close() error {
	ep.once.Do(func() { close(ep.stop) })
	<-ep.done
	return ep.err
}
