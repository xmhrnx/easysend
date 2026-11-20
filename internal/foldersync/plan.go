package foldersync

import "sort"

// actionKind is what pulling from a peer does with one file.
type actionKind int

const (
	// agree records that both devices have the same version.
	agree actionKind = iota
	// fetch downloads the peer's version over this device's.
	fetch
	// remove deletes this device's copy, as the peer did.
	remove
	// keepBoth renames this device's copy to a conflict copy, then
	// downloads the peer's version in its place.
	keepBoth
)

type action struct {
	kind   actionKind
	path   string
	remote Entry
}

// plan compares this device's index with a peer's, against base, the
// versions the two last agreed on by path, and says what to do. Files
// only the peer changed are taken from it; files only this device
// changed are left for the peer to take; files both changed are
// resolved the same way on both sides, so they end up equal:
//
//   - a change beats a deletion;
//   - of two edits, the later one (Changed, then By) wins, and the
//     losing device keeps its version as a conflict copy, which then
//     syncs like any new file.
func plan(local, remote Index, base map[string]string) []action {
	var acts []action
	for p, r := range remote {
		rv := r.version()
		l, had := local[p]
		lv := ""
		if had {
			lv = l.version()
		}
		switch {
		case lv == rv:
			if base[p] != rv {
				acts = append(acts, action{agree, p, r})
			}
		case !had:
			// Never seen here: take it, or just note the deletion.
			if rv == deletedVersion {
				acts = append(acts, action{agree, p, r})
			} else {
				acts = append(acts, action{fetch, p, r})
			}
		case lv == base[p]:
			// Only the peer changed it.
			if rv == deletedVersion {
				acts = append(acts, action{remove, p, r})
			} else {
				acts = append(acts, action{fetch, p, r})
			}
		case rv == base[p]:
			// Only this device changed it: the peer takes it from here.
		case rv == deletedVersion:
			// Changed here, deleted there: the change wins (the peer
			// fetches it).
		case lv == deletedVersion:
			// Deleted here, changed there: the change wins.
			acts = append(acts, action{fetch, p, r})
		case newer(r, l):
			acts = append(acts, action{keepBoth, p, r})
		default:
			// Both edited and this device's is newer: the peer keeps its
			// own as a conflict copy and fetches this one.
		}
	}
	sort.Slice(acts, func(i, j int) bool { return acts[i].path < acts[j].path })
	return acts
}

// newer reports whether a is the later of two edits made apart.
func newer(a, b Entry) bool {
	if a.Changed != b.Changed {
		return a.Changed > b.Changed
	}
	return a.By > b.By
}
