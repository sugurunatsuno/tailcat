package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"mime"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/sugurunatsuno/tailcat/internal/protocol"
	"github.com/tailscale/tailcat"
)

const (
	filecatPort = 7331
	receiverURL = "https://sugurunatsuno.github.io/tailcat/"
)

func main() {
	flag.Usage = func() { fmt.Fprintln(os.Stderr, "usage: filecat <file>") }
	noQR := flag.Bool("no-qr", false, "do not print a QR code")
	urlOnly := flag.Bool("url-only", false, "print only the receiver URL")
	receiver := flag.String("receiver", receiverURL, "receiver page URL")
	timeout := flag.Duration("timeout", 15*time.Minute, "wait timeout; 0 waits forever")
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	path := flag.Arg(0)
	info, err := os.Stat(path)
	if err != nil {
		fatal(err)
	}
	if !info.Mode().IsRegular() {
		fatal(fmt.Errorf("not a regular file: %s", path))
	}

	var busy atomic.Bool
	done := make(chan struct{})
	s := &tailcat.Server{OnTCP: func(port uint16) func(net.Conn) {
		if port != filecatPort || !busy.CompareAndSwap(false, true) {
			return nil
		}
		return func(conn net.Conn) {
			defer busy.Store(false)
			if serve(conn, path, info) {
				close(done)
			}
		}
	}}
	if err := s.Start(); err != nil {
		fatal(err)
	}
	defer s.Close()

	url := fmt.Sprintf("%s#v=1&tc=%s", *receiver, s.TailcatAddr())
	if *urlOnly {
		fmt.Println(url)
	} else {
		fmt.Printf("\n%s  %.1f MB\n\n%s\n\n", filepath.Base(path), float64(info.Size())/1e6, url)
		if !*noQR {
			qrterminal.GenerateWithConfig(url, qrterminal.Config{Level: qrterminal.M, Writer: os.Stdout, BlackChar: "██", WhiteChar: "  ", QuietZone: 2})
		}
		fmt.Println("\nWaiting for receiver...")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if *timeout == 0 {
		select {
		case <-ctx.Done():
		case <-done:
			fmt.Println("Done")
		}
		return
	}
	timer := time.NewTimer(*timeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-done:
		fmt.Println("Done")
	case <-timer.C:
		fmt.Fprintln(os.Stderr, "filecat: timeout")
	}
}

func serve(conn net.Conn, path string, info os.FileInfo) bool {
	defer conn.Close()
	name := filepath.Base(path)
	mimeType := mime.TypeByExtension(filepath.Ext(name))
	if err := protocol.Write(conn, protocol.Message{Type: "hello", Version: protocol.Version, Name: name, Size: info.Size(), MIME: mimeType}); err != nil {
		return false
	}
	message, err := protocol.Read(conn)
	if err != nil || message.Type != "get" {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	if _, err = io.CopyN(conn, f, info.Size()); err != nil {
		return false
	}
	return protocol.Write(conn, protocol.Message{Type: "done"}) == nil
}

func fatal(err error) { fmt.Fprintln(os.Stderr, "filecat:", err); os.Exit(1) }
