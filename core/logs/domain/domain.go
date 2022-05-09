package domain

import (
	"bufio"
	"os"
	"sync"
)

type Store struct {
	*os.File
	mu  sync.Mutex
	buf *bufio.Writer
}
