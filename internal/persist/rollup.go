package persist

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/zyvorai/shukra/internal/state"
)

const rollupFile = "snapshots.jsonl"

// Rollup keeps the coarse counter snapshots a verdict for a past time stands on. It is a state.RollupStore
// over a rolling JSON-lines file, so how far back it goes is set by size (DefaultMaxBytes, twice with the
// rolled file), not by time: about a day at ten VMs, less for more.
type Rollup struct {
	log  *Log
	path string
}

// OpenRollup opens the snapshot log under dir.
func OpenRollup(dir string) (*Rollup, error) {
	path := dir + string(os.PathSeparator) + rollupFile
	l, err := OpenLog(path, DefaultMaxBytes)
	if err != nil {
		return nil, err
	}
	return &Rollup{log: l, path: path}, nil
}

// Close closes the log.
func (r *Rollup) Close() error { return r.log.Close() }

// Append implements state.RollupStore.
func (r *Rollup) Append(s state.RollupSnap) error { return r.log.Append(s) }

// place is where one snapshot sits in the files, found without keeping the snapshot.
type place struct {
	at   time.Time
	file string
	off  int64
}

// scan lists every stored snapshot's time and position, oldest first, path.1 before path. A record that
// does not decode (a line torn by a crash) is skipped.
func (r *Rollup) scan() ([]place, error) {
	var out []place
	var rerr error
	for _, p := range []string{r.path + ".1", r.path} {
		f, err := os.Open(p)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			rerr = err
			continue
		}
		// Read line by line: Log.open ends a torn line with a newline and appends after it, so a bad
		// record in the middle of the file must be skipped, not treated as the end.
		br := bufio.NewReaderSize(f, 1<<20)
		var off int64
		for {
			line, err := br.ReadBytes('\n')
			if len(line) > 0 {
				var head struct {
					At time.Time `json:"at"`
				}
				if json.Unmarshal(line, &head) == nil {
					out = append(out, place{at: head.At, file: p, off: off})
				}
				off += int64(len(line))
			}
			if err != nil {
				break
			}
		}
		f.Close()
	}
	return out, rerr
}

func (r *Rollup) load(p place) (*state.RollupSnap, error) {
	f, err := os.Open(p.file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := f.Seek(p.off, io.SeekStart); err != nil {
		return nil, err
	}
	var s state.RollupSnap
	if err := json.NewDecoder(bufio.NewReader(f)).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Around implements state.RollupStore. It reads the times of all snapshots and then only the two it needs.
func (r *Rollup) Around(at time.Time, window time.Duration) (cur, base *state.RollupSnap, err error) {
	all, err := r.scan()
	if err != nil {
		return nil, nil, err
	}
	ci := -1
	for i, p := range all {
		if !p.at.After(at) {
			ci = i
		}
	}
	if ci < 0 {
		return nil, nil, nil
	}
	bi := -1
	for i := ci - 1; i >= 0; i-- {
		if !all[i].at.After(all[ci].at.Add(-window)) {
			bi = i
			break
		}
	}
	if cur, err = r.load(all[ci]); err != nil {
		return nil, nil, err
	}
	if bi >= 0 {
		if base, err = r.load(all[bi]); err != nil {
			return nil, nil, err
		}
	}
	return cur, base, nil
}
