package server

import stdos "os"

var os = struct {
	ErrNotExist error
}{
	ErrNotExist: stdos.ErrNotExist,
}
