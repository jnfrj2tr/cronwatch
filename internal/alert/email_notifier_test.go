package alert

import (
	"fmt"
	"io"
	"net"
	"net/textproto"
	"testing"
)

func TestEmailNotifier_Send_NoRecipients(t *testing.T) {
	n := NewEmailNotifier(EmailConfig{
		Host: "localhost",
		Port: 2525,
		From: "cronwatch@example.com",
		To:   []string{},
	})
	err := n.Send("test alert")
	if err == nil {
		t.Fatal("expected error for empty recipients, got nil")
	}
}

func TestEmailNotifier_Send_BadHost(t *testing.T) {
	n := NewEmailNotifier(EmailConfig{
		Host: "127.0.0.1",
		Port: 19999,
		From: "cronwatch@example.com",
		To:   []string{"ops@example.com"},
	})
	err := n.Send("test alert")
	if err == nil {
		t.Fatal("expected error when SMTP host unreachable, got nil")
	}
}

func stubSMTPServer(t *testing.T) (host string, port int, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		tc := textproto.NewConn(conn)
		_ = tc.PrintfLine("220 stub ready")
		for {
			line, err := tc.ReadLine()
			if err == io.EOF || err != nil {
				return
			}
			switch {
			case len(line) >= 4 && (line[:4] == "EHLO" || line[:4] == "HELO"):
				_ = tc.PrintfLine("250 ok")
			case len(line) >= 4 && (line[:4] == "MAIL" || line[:4] == "RCPT"):
				_ = tc.PrintfLine("250 ok")
			case line == "DATA":
				_ = tc.PrintfLine("354 go ahead")
			case line == ".":
				_ = tc.PrintfLine("250 queued")
			case line == "QUIT":
				_ = tc.PrintfLine("221 bye")
				return
			}
		}
	}()
	var h string
	var p int
	_, err = fmt.Sscanf(ln.Addr().String(), "%[^:]:%d", &h, &p)
	if err != nil {
		t.Fatalf("parse addr: %v", err)
	}
	return h, p, func() { ln.Close() }
}

func TestEmailNotifier_Send_Success(t *testing.T) {
	h, p, stop := stubSMTPServer(t)
	defer stop()

	n := NewEmailNotifier(EmailConfig{
		Host: h,
		Port: p,
		From: "cronwatch@example.com",
		To:   []string{"ops@example.com"},
	})
	if err := n.Send("backup job missed"); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}
}
