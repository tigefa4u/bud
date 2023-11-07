package slots

import (
	"bytes"
	"io"
	"sync"
)

func New() *Slots {
	return &Slots{
		nth:    0,
		reader: bytes.NewBuffer(nil),
		inners: make(map[string]*pipe),
		pipe:   newPipe(),
	}
}

type Pipe interface {
	io.ReadWriter
}

type Slots struct {
	nth     int
	reader  io.Reader
	pipe    *pipe
	mu      sync.Mutex
	inners  map[string]*pipe
	closers []io.Closer
}

var _ io.ReadWriteCloser = (*Slots)(nil)

func (s *Slots) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

func (s *Slots) Write(p []byte) (n int, err error) {
	return s.pipe.Write(p)
}

func (s *Slots) Close() error {
	for _, closer := range s.closers {
		closer.Close()
	}
	return s.pipe.Close()
}

func (s *Slots) Next() *Slots {
	pipe := newPipe()
	return &Slots{
		nth:    s.nth + 1,
		reader: s.pipe,
		inners: s.inners,
		pipe:   pipe,
	}
}

func (s *Slots) Pipe(name string) Pipe {
	s.mu.Lock()
	defer s.mu.Unlock()
	pipe, ok := s.inners[name]
	if !ok {
		pipe := newPipe()
		s.inners[name] = pipe
		s.closers = append(s.closers, pipe)
		return pipe
	}
	return pipe
}
