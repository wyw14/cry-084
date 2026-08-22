package files

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/local/cry-084/internal/application"
	"github.com/local/cry-084/internal/domain/shared"
)

type Local struct {
	root     string
	maxBytes int64
	metadata map[shared.ID]application.FileMetadata
}

func NewLocal(root string, maxBytes int64) (*Local, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absolute, 0o750); err != nil {
		return nil, err
	}
	return &Local{root: absolute, maxBytes: maxBytes, metadata: map[shared.ID]application.FileMetadata{}}, nil
}

func (s *Local) Put(ctx context.Context, meta application.FileMetadata, reader io.Reader) error {
	policy := uploadPolicy{root: s.root, maxBytes: s.maxBytes}
	if err := policy.validateMetadata(meta); err != nil {
		return err
	}
	target := filepath.Join(s.root, string(meta.ID))
	if !strings.HasPrefix(target, s.root+string(os.PathSeparator)) {
		return shared.ErrForbidden
	}
	temp := target + ".tmp"
	file, err := os.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(file, hash), io.LimitReader(reader, s.maxBytes+1))
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(temp)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(temp)
		return closeErr
	}
	if written != meta.Size || written > s.maxBytes || hex.EncodeToString(hash.Sum(nil)) != meta.Digest {
		_ = os.Remove(temp)
		return fmt.Errorf("%w: file size or digest mismatch", shared.ErrValidation)
	}
	select {
	case <-ctx.Done():
		_ = os.Remove(temp)
		return ctx.Err()
	default:
	}
	if err := os.Rename(temp, target); err != nil {
		_ = os.Remove(temp)
		return err
	}
	s.metadata[meta.ID] = meta
	return nil
}

type uploadPolicy struct {
	root     string
	maxBytes int64
}

func (p uploadPolicy) validateMetadata(meta application.FileMetadata) error {
	if err := p.validateIdentity(meta); err != nil {
		return err
	}
	if err := p.validateSize(meta); err != nil {
		return err
	}
	return p.validateName(meta)
}

func (p uploadPolicy) validateIdentity(meta application.FileMetadata) error {
	if meta.ID == "" || meta.MallID == "" {
		return fmt.Errorf("%w: file identity is required", shared.ErrValidation)
	}
	return nil
}

func (p uploadPolicy) validateSize(meta application.FileMetadata) error {
	if meta.Size < 1 || meta.Size > p.maxBytes {
		return fmt.Errorf("%w: invalid file size", shared.ErrValidation)
	}
	return nil
}

func (p uploadPolicy) validateName(meta application.FileMetadata) error {
	if strings.Contains(meta.Name, "..") || strings.ContainsAny(meta.Name, `/\\`) {
		return fmt.Errorf("%w: unsafe file name", shared.ErrValidation)
	}
	return nil
}

func (p uploadPolicy) allowedMIME(value string) bool {
	switch value {
	case "image/jpeg", "image/png", "application/pdf":
		return true
	default:
		return false
	}
}

func (p uploadPolicy) target(id shared.ID) (string, error) {
	target := filepath.Join(p.root, string(id))
	if !strings.HasPrefix(target, p.root+string(os.PathSeparator)) {
		return "", shared.ErrForbidden
	}
	return target, nil
}

func (p uploadPolicy) temporary(id shared.ID) (string, error) {
	target, err := p.target(id)
	if err != nil {
		return "", err
	}
	return target + ".tmp", nil
}

func (s *Local) Open(ctx context.Context, id shared.ID) (application.FileMetadata, io.ReadCloser, error) {
	meta, ok := s.metadata[id]
	if !ok {
		return application.FileMetadata{}, nil, shared.ErrNotFound
	}
	select {
	case <-ctx.Done():
		return application.FileMetadata{}, nil, ctx.Err()
	default:
	}
	file, err := os.Open(filepath.Join(s.root, string(id)))
	return meta, file, err
}
