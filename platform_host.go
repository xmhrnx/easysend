package main

import (
	"context"

	"github.com/xmhrnx/giode"

	"easysend/ui/platform"
)

// publisher moves received files to the system Downloads where the
// platform keeps them apart from the app (Android).
func publisher(app *giode.App) func(context.Context, string) (string, error) {
	if !platform.ReceivesToDownloads {
		return nil
	}
	return func(ctx context.Context, path string) (string, error) {
		return app.SaveToDownloads(ctx, path, "EasySend")
	}
}

// deviceModel names the hardware for other devices' lists.
func deviceModel() string {
	d := giode.Device()
	if d.Model != "" {
		return d.Model
	}
	switch d.OS {
	case "darwin":
		return "Mac"
	case "windows":
		return "Windows PC"
	case "linux":
		return "Linux PC"
	case "android":
		return "Android"
	case "ios":
		return "iPhone"
	}
	return "Computer"
}
