//go:build js && wasm

package main

import (
	"context"
	"errors"
	"github.com/tailscale/tailcat"
	"net"
	"syscall/js"
	"time"
)

func main() {
	js.Global().Set("tailcatDial", js.FuncOf(dial))
	if ready := js.Global().Get("onTailcatReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}
	select {}
}
func dial(_ js.Value, args []js.Value) any {
	if len(args) != 1 {
		return js.Global().Get("Promise").Call("reject", "options required")
	}
	return promise(func() (any, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		c := &tailcat.Client{Server: tailcat.Addr(args[0].Get("addr").String())}
		conn, err := c.DialTCPPort(ctx, 7331)
		if err != nil {
			c.Close()
			return nil, err
		}
		return wrap(conn, c), nil
	})
}
func wrap(conn net.Conn, client *tailcat.Client) js.Value {
	read := js.FuncOf(func(js.Value, []js.Value) any {
		return promise(func() (any, error) {
			b := make([]byte, 32*1024)
			n, err := conn.Read(b)
			if err != nil {
				return nil, err
			}
			out := js.Global().Get("Uint8Array").New(n)
			js.CopyBytesToJS(out, b[:n])
			return out, nil
		})
	})
	write := js.FuncOf(func(_ js.Value, args []js.Value) any {
		return promise(func() (any, error) {
			if len(args) != 1 {
				return nil, errors.New("bytes required")
			}
			b := make([]byte, args[0].Length())
			js.CopyBytesToGo(b, args[0])
			_, err := conn.Write(b)
			return nil, err
		})
	})
	close := js.FuncOf(func(js.Value, []js.Value) any { conn.Close(); client.Close(); return nil })
	return js.ValueOf(map[string]any{"read": read, "write": write, "close": close})
}
func promise(fn func() (any, error)) js.Value {
	return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, args []js.Value) any {
		go func() {
			v, err := fn()
			if err != nil {
				args[0].Invoke(err.Error())
			} else {
				args[1].Invoke(v)
			}
		}()
		return nil
	}))
}
