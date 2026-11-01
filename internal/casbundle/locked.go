package casbundle

import (
	"io"
	"os"
)

type lockedBundle struct {
	handles    []*os.File
	files      map[string]*os.File
	executable string
}

func (b *lockedBundle) close() {
	for i := len(b.handles) - 1; i >= 0; i-- {
		_ = b.handles[i].Close()
	}
}
func (b *lockedBundle) read(name string, limit int64) ([]byte, error) {
	f := b.files[name]
	if f == nil {
		return nil, ErrInvalid
	}
	info, err := f.Stat()
	if err != nil || info.Size() < 1 || info.Size() > limit {
		return nil, ErrInvalid
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(data)) != info.Size() {
		return nil, ErrInvalid
	}
	return data, nil
}
