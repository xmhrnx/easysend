package wormhole

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	ww "github.com/psanford/wormhole-william/wormhole"

	"easysend/internal/item"
)

// Send offers items and returns the code the receiver types. The returned
// channel yields the outcome once the receiver took them (nil) or the
// transfer failed.
func Send(ctx context.Context, o Options, items []item.Item, progress func(sent, total int64)) (string, <-chan error, error) {
	if len(items) == 0 {
		return "", nil, errors.New("nothing to send")
	}
	c := o.client()
	opt := ww.WithProgress(progress)
	var (
		code string
		res  chan ww.SendResult
		err  error
		file *os.File
	)
	switch {
	case len(items) == 1 && items[0].Kind == item.Text:
		code, res, err = c.SendText(ctx, items[0].Text, opt)
	case len(items) == 1:
		file, err = os.Open(items[0].Path)
		if err != nil {
			return "", nil, err
		}
		code, res, err = c.SendFile(ctx, items[0].BaseName(), file, opt)
	default:
		code, res, err = c.SendDirectory(ctx, folderName(items), entries(items), opt)
	}
	if err != nil {
		if file != nil {
			file.Close()
		}
		return "", nil, err
	}
	out := make(chan error, 1)
	go func() {
		if file != nil {
			defer file.Close()
		}
		select {
		case r := <-res:
			if !r.OK && r.Error == nil {
				r.Error = errors.New("transfer failed")
			}
			out <- r.Error
		case <-ctx.Done():
			out <- ctx.Err()
		}
	}()
	return code, out, nil
}

func folderName(items []item.Item) string {
	group := items[0].Group
	for _, it := range items {
		if it.Group != group {
			group = ""
			break
		}
	}
	if group != "" {
		return group
	}
	return fmt.Sprintf("EasySend (%d files)", len(items))
}

// entries lists items for a directory transfer. Items from one folder
// drop the folder's own name, which becomes the directory's.
func entries(items []item.Item) []ww.DirectoryEntry {
	single := folderName(items) == items[0].Group && items[0].Group != ""
	out := make([]ww.DirectoryEntry, 0, len(items))
	for _, it := range items {
		name := it.Name
		if single {
			name = strings.TrimPrefix(name, it.Group+"/")
		}
		it := it
		out = append(out, ww.DirectoryEntry{Path: name, Mode: 0o644, Reader: it.Open})
	}
	return out
}
