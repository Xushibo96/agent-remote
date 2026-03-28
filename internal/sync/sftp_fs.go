package sync

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"agent-remote/internal/interfaces"
	"agent-remote/internal/model"
	"github.com/pkg/sftp"
)

type SFTPFilesystem struct {
	root   string
	client *sftp.Client
}

func NewSFTPFilesystem(ctx context.Context, connections interfaces.ConnectionManager, target model.ResolvedTarget, root string) (*SFTPFilesystem, func() error, error) {
	client, err := connections.GetSSHClient(ctx, target)
	if err != nil {
		return nil, nil, err
	}
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, nil, err
	}
	return &SFTPFilesystem{root: root, client: sftpClient}, sftpClient.Close, nil
}

func (s *SFTPFilesystem) join(name string) string {
	clean := path.Clean("/" + strings.TrimSpace(name))
	if s.root == "" {
		return clean
	}
	return path.Join(s.root, clean)
}

func (s *SFTPFilesystem) Stat(name string) (os.FileInfo, error) {
	return s.client.Stat(s.join(name))
}

func (s *SFTPFilesystem) ReadDir(name string) ([]os.DirEntry, error) {
	entries, err := s.client.ReadDir(s.join(name))
	if err != nil {
		return nil, err
	}
	out := make([]os.DirEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, fs.FileInfoToDirEntry(entry))
	}
	return out, nil
}

func (s *SFTPFilesystem) ReadFile(name string) ([]byte, error) {
	file, err := s.client.Open(s.join(name))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func (s *SFTPFilesystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	if err := s.client.MkdirAll(path.Dir(s.join(name))); err != nil {
		return err
	}
	file, err := s.client.OpenFile(s.join(name), os.O_CREATE|os.O_RDWR|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	return file.Chmod(perm)
}

func (s *SFTPFilesystem) MkdirAll(name string, _ os.FileMode) error {
	return s.client.MkdirAll(s.join(name))
}

func (s *SFTPFilesystem) RemoveAll(name string) error {
	walker := s.client.Walk(s.join(name))
	paths := make([]string, 0)
	for walker.Step() {
		if walker.Err() != nil {
			return walker.Err()
		}
		paths = append(paths, walker.Path())
	}
	for i := len(paths) - 1; i >= 0; i-- {
		info, err := s.client.Stat(paths[i])
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := s.client.RemoveDirectory(paths[i]); err != nil {
				return err
			}
			continue
		}
		if err := s.client.Remove(paths[i]); err != nil {
			return err
		}
	}
	return nil
}
