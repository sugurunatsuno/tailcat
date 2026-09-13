package protocol

import (
	"bytes"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	var b bytes.Buffer
	want := Message{Type: "hello", Version: 1, Name: "photo.jpg", Size: 42, MIME: "image/jpeg"}
	if err := Write(&b, want); err != nil { t.Fatal(err) }
	got, err := Read(&b); if err != nil { t.Fatal(err) }
	if got != want { t.Fatalf("got %+v, want %+v", got, want) }
}
