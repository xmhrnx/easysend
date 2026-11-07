package transfer

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const historyLimit = 300

func (r *Registry) save() {
	list := r.List()
	var keep []Transfer
	for _, t := range list {
		if t.Status.Finished() {
			keep = append(keep, t)
		}
	}
	if len(keep) > historyLimit {
		keep = keep[:historyLimit]
	}
	data, err := json.Marshal(keep)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(r.path), 0o700)
	_ = os.WriteFile(r.path, data, 0o600)
}
