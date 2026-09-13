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

	"github.com/sugurunatsuno/tailcat/internal/protocol"
	"github.com/tailscale/tailcat"
)

const (
	filecatPort = 7331
	receiverURL = "https://username.github.io/filecat/"
)

func main() {
	flag.Usage = func() { fmt.Fprintln(os.Stderr, "usage: filecat <file>") }
	flag.Parse()
	if flag.NArg() != 1 { flag.Usage(); os.Exit(2) }

	path := flag.Arg(0)
	info, err := os.Stat(path)
	if err != nil { fatal(err) }
	if !info.Mode().IsRegular() { fatal(fmt.Errorf("not a regular file: %s", path)) }

	var busy atomic.Bool
	s := &tailcat.Server{OnTCP: func(port uint16) func(net.Conn) {
		if port != filecatPort || !busy.CompareAndSwap(false, true) { return nil }
		return func(conn net.Conn) { defer busy.Store(false); serve(conn, path, info) }
	}}
	if err := s.Start(); err != nil { fatal(err) }
	defer s.Close()

	url := fmt.Sprintf("%s#v=1&tc=%s", receiverURL, s.TailcatAddr())
	fmt.Printf("\n%s  %.1f MB\n\n%s\n\nWaiting for receiver...\n", filepath.Base(path), float64(info.Size())/1e6, url)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	<-ctx.Done()
}

func serve(conn net.Conn, path string, info os.FileInfo) {
	defer conn.Close()
	name := filepath.Base(path)
	mimeType := mime.TypeByExtension(filepath.Ext(name))
	if err := protocol.Write(conn, protocol.Message{Type: "hello", Version: protocol.Version, Name: name, Size: info.Size(), MIME: mimeType}); err != nil { return }
	message, err := protocol.Read(conn)
	if err != nil || message.Type != "get" { return }
	f, err := os.Open(path); if err != nil { return }; defer f.Close()
	if _, err = io.CopyN(conn, f, info.Size()); err != nil { return }
	_ = protocol.Write(conn, protocol.Message{Type: "done"})
}

func fatal(err error) { fmt.Fprintln(os.Stderr, "filecat:", err); os.Exit(1) }
