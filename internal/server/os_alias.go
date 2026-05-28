package server

import stdos "os"

var oos = struct {
	ErrNotExist error
}{
	ErrNotExist: stdos.ErrNotExist,
}
