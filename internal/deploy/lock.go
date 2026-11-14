package deploy

import (
	"errors"
	"os"
)

var ErrAlreadyLocked = errors.New("deployment already locked")

type Lock struct {
	path string
	file *os.File
}

func NewLock(path string) *Lock {
	return &Lock{path: path}
}

func (l *Lock) Acquire() error {
	if l.file != nil {
		return ErrAlreadyLocked
	}
	file, err := os.OpenFile(l.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrAlreadyLocked
		}
		return err
	}
	l.file = file
	return nil
}

func (l *Lock) Release() error {
	if l.file == nil {
		return nil
	}
	if err := l.file.Close(); err != nil {
		return err
	}
	l.file = nil
	return os.Remove(l.path)
}
