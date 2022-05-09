package usecases

import (
	"encoding/json"
	"os"
	"time"

	repository "github.com/mindaugasrmn/distributed-cowboys/core/logs/repository"
)

type Row struct {
	Time time.Time
	Msg  string
}

type logUsecases struct {
	Dir   string
	store repository.Store
}

type LogUsecases interface {
	Append(record string) (uint64, error)
	Close() error
	Remove() error
}

func NewLog(store repository.Store) LogUsecases {
	return &logUsecases{
		store: store,
	}
}

func (l *logUsecases) Append(record string) (uint64, error) {
	row := Row{time.Now(), record}
	b, err := json.Marshal(row)
	if err != nil {
		return 0, err
	}
	off, _, err := l.store.AddLine(b)
	if err != nil {
		return 0, err
	}

	return off, err
}

func (l *logUsecases) Close() error {
	if err := l.store.Close(); err != nil {
		return err
	}
	return nil
}

func (l *logUsecases) Remove() error {
	if err := l.Close(); err != nil {
		return err
	}
	return os.RemoveAll(l.Dir)
}
