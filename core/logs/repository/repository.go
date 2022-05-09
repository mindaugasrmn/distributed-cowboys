package repository

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

const (
	lenWidth = 8
)

type Store interface {
	AddLine(p []byte) (n uint64, pos uint64, err error)
	Close() error
}

type store struct {
	*os.File
	mu  sync.Mutex
	buf *bufio.Writer
}

func NewStore(
	f *os.File,
) Store {
	return &store{
		File: f,
		buf:  bufio.NewWriter(f),
	}

}

func (s *store) AddLine(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	written, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}
	fmt.Fprintf(s.buf, "\n")
	err = s.buf.Flush()
	if err != nil {
		return 0, 0, err
	}
	return uint64(written), pos, nil
}

func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}
