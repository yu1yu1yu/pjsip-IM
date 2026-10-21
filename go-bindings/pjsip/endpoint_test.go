package pjsip

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

func unusedPort(t *testing.T) int {
	t.Helper()
	conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := conn.LocalAddr().(*net.UDPAddr).Port
	conn.Close()
	return port
}

func TestLifecycle(t *testing.T) {
	if Version() == "" {
		t.Fatal("empty native version")
	}
	if _, err := New(Config{Port: 0, Username: "alice"}); err == nil {
		t.Fatal("accepted invalid port")
	}
	if _, err := New(Config{Port: 5060, Username: "bad\r\nname"}); err == nil {
		t.Fatal("accepted invalid username")
	}
	// A failed transport creation must release both the native state and lock.
	conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := conn.LocalAddr().(*net.UDPAddr).Port
	if ep, err := New(Config{Port: port, Username: "occupied"}); err == nil {
		ep.Close()
		t.Fatal("accepted occupied port")
	}
	conn.Close()
	for i := 0; i < 2; i++ {
		ep, err := New(Config{Port: port, Username: "alice"})
		if err != nil {
			t.Fatal(err)
		}
		if other, err := New(Config{Port: unusedPort(t), Username: "bob"}); err == nil {
			other.Close()
			t.Fatal("accepted second endpoint")
		}
		if _, err := ep.Send("bad\r\nuri", "hello"); err == nil {
			t.Fatal("accepted invalid URI")
		}
		if _, err := ep.Send("sip:bob@127.0.0.1:5070", strings.Repeat("x", 2049)); err == nil {
			t.Fatal("accepted oversized body")
		}
		if err := ep.Close(); err != nil {
			t.Fatal(err)
		}
		if err := ep.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := ep.Send("sip:bob@127.0.0.1:5070", "closed"); err != ErrClosed {
			t.Fatalf("send after close: %v", err)
		}
	}
}

func TestPeerHelper(t *testing.T) {
	if os.Getenv("PJGO_TEST_PEER") != "1" {
		return
	}
	port, _ := strconv.Atoi(os.Getenv("PJGO_TEST_PORT"))
	ep, err := New(Config{Port: port, Username: "bob"})
	if err != nil {
		t.Fatal(err)
	}
	defer ep.Close()
	fmt.Println("PJGO_PEER_READY")
	timeout := time.After(15 * time.Second)
	received := false
	for {
		select {
		case event, ok := <-ep.Events():
			if !ok {
				t.Fatalf("peer stopped: %v", ep.Err())
			}
			if event.Kind == "message" {
				if event.Body != "你好 Bob 📨" {
					t.Fatalf("wrong request: %q", event.Body)
				}
				if !strings.HasPrefix(event.ContentType, "text/plain") {
					t.Fatalf("wrong MIME: %s", event.ContentType)
				}
				received = true
				if _, err := ep.Send(event.ReplyTo, "收到，Alice！"); err != nil {
					t.Fatal(err)
				}
			}
			if event.Kind == "status" {
				if event.Code != 200 || !received {
					t.Fatalf("wrong reply status: %+v", event)
				}
				return
			}
		case <-timeout:
			t.Fatal("peer timed out")
		}
	}
}

func TestChineseRoundTrip(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	peerPort := unusedPort(t)
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPeerHelper$", "-test.v")
	cmd.Env = append(os.Environ(), "PJGO_TEST_PEER=1", fmt.Sprintf("PJGO_TEST_PORT=%d", peerPort))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	defer func() {
		if !waited {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()
	ready := make(chan bool, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		found := false
		for scanner.Scan() {
			fmt.Fprintln(os.Stderr, "peer:", scanner.Text())
			if scanner.Text() == "PJGO_PEER_READY" {
				ready <- true
				found = true
			}
		}
		if !found {
			ready <- false
		}
	}()
	select {
	case ok := <-ready:
		if !ok {
			t.Fatal("peer exited before ready")
		}
	case <-ctx.Done():
		t.Fatal("peer startup timed out")
	}
	ep, err := New(Config{Port: unusedPort(t), Username: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	defer ep.Close()
	id, err := ep.Send(fmt.Sprintf("sip:bob@127.0.0.1:%d", peerPort), "你好 Bob 📨")
	if err != nil {
		t.Fatal(err)
	}
	received, accepted := false, false
	for !received || !accepted {
		select {
		case event, ok := <-ep.Events():
			if !ok {
				t.Fatalf("endpoint stopped: %v", ep.Err())
			}
			t.Logf("alice event: %+v", event)
			if event.Kind == "message" {
				if event.Body != "收到，Alice！" {
					t.Fatalf("wrong reply: %q", event.Body)
				}
				received = true
			}
			if event.Kind == "status" {
				if event.ID != id || event.Code != 200 {
					t.Fatalf("wrong delivery: %+v", event)
				}
				accepted = true
			}
		case <-ctx.Done():
			t.Fatal("round trip timed out")
		}
	}
	err = cmd.Wait()
	waited = true
	if err != nil {
		t.Fatalf("peer process failed: %v", err)
	}
}
