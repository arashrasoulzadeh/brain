package plugin

import (
	"bufio"
	"encoding/json"
	"errors"
)

type workerHostCall struct {
	ID        uint64          `json:"id"`
	Operation string          `json:"operation"`
	Input     json.RawMessage `json:"input"`
}

type workerHostReply struct {
	ID     uint64          `json:"id"`
	Output json.RawMessage `json:"output,omitempty"`
	Error  string          `json:"error,omitempty"`
}

func readWorkerLine(reader *bufio.Reader, limit int) ([]byte, error) {
	var line []byte
	for {
		part, err := reader.ReadSlice('\n')
		if len(line)+len(part) > limit {
			return nil, ErrInvalidRPCMessage
		}
		line = append(line, part...)
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		return line, err
	}
}
