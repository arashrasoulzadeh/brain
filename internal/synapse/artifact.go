package synapse

import (
	"bytes"
	"context"
	"errors"

	"brain/internal/execution"
)

const ArtifactChunkBytes = 64 << 10

var ErrInvalidChunk = errors.New("invalid synapse artifact chunk")

type Chunk struct {
	Lease         Lease
	Offset, Total int64
	Data          []byte
	Final         bool
}

// StreamArtifact reads only the artifact associated with a live assignment.
// A single stream writer must serialize send with other outbound messages.
func StreamArtifact(ctx context.Context, session *Session, catalog *execution.Catalog, lease Lease, send func(Chunk) error) error {
	if session == nil || catalog == nil || send == nil {
		return ErrProtocol
	}
	artifact, err := session.ArtifactFor(lease)
	if err != nil {
		return err
	}
	data, err := catalog.Fetch(artifact)
	if err != nil {
		return err
	}
	for offset := 0; offset < len(data); {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := session.ArtifactFor(lease); err != nil {
			return err
		}
		end := min(offset+ArtifactChunkBytes, len(data))
		chunk := Chunk{Lease: lease, Offset: int64(offset), Total: int64(len(data)), Data: bytes.Clone(data[offset:end]), Final: end == len(data)}
		if err := send(chunk); err != nil {
			return err
		}
		offset = end
	}
	return nil
}

// Assembler is owned by a single stream receiver and never exposes partial
// artifact bytes. An invalid chunk poisons the transfer and drops its buffer.
type Assembler struct {
	lease            Lease
	artifact         execution.Artifact
	limit, total     int64
	data             []byte
	complete, failed bool
}

func NewAssembler(lease Lease, artifact execution.Artifact, maxBytes int64) (*Assembler, error) {
	if !validID(lease.SessionID) || lease.Attempt == 0 || !validID(lease.ExecutionID) || maxBytes <= 0 || maxBytes > execution.MaxBundleBytes {
		return nil, ErrInvalidChunk
	}
	return &Assembler{lease: lease, artifact: artifact, limit: maxBytes}, nil
}
func (a *Assembler) reject(err error) error { a.failed = true; a.data = nil; return err }
func (a *Assembler) Push(chunk Chunk) error {
	if a.failed || a.complete || chunk.Lease != a.lease || chunk.Offset != int64(len(a.data)) || chunk.Total <= 0 || chunk.Total > a.limit || (a.total != 0 && a.total != chunk.Total) || len(chunk.Data) == 0 || len(chunk.Data) > ArtifactChunkBytes || int64(len(chunk.Data)) > chunk.Total-chunk.Offset {
		return a.reject(ErrInvalidChunk)
	}
	next := chunk.Offset + int64(len(chunk.Data))
	if chunk.Final != (next == chunk.Total) {
		return a.reject(ErrInvalidChunk)
	}
	a.total = chunk.Total
	a.data = append(a.data, chunk.Data...)
	if chunk.Final {
		if err := a.artifact.Verify(a.data); err != nil {
			return a.reject(err)
		}
		a.complete = true
	}
	return nil
}
func (a *Assembler) Finish() ([]byte, error) {
	if a.failed || !a.complete {
		return nil, ErrInvalidChunk
	}
	data := a.data
	a.data = nil
	a.failed = true
	return data, nil
}
