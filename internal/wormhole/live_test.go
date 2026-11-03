//go:build live

package wormhole

import (
	"context"
	"testing"
	"time"

	"easysend/internal/item"
)

func TestLiveText(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	code, res, err := Send(ctx, Options{}, []item.Item{item.FromText("hi there")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	of, err := Receive(ctx, Options{}, code)
	if err != nil {
		t.Fatal(err)
	}
	txt, err := of.ReadText()
	if err != nil || txt != "hi there" {
		t.Fatal(txt, err)
	}
	if err := <-res; err != nil {
		t.Fatal(err)
	}
}
