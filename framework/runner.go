package framework

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type ImplementedScripts int

const NoneImplemented ImplementedScripts = 0

const (
	Check ImplementedScripts = 1 << iota
	In
	Out
)

type ResourceType[S any, V any, G any, P any] struct {
	stdout io.Writer
	stdin  io.Reader
	check  Checker[S, V]
	in     Getter[S, V, G]
	out    Putter[S, V, P]
}

type ResourceTypeOption[SourceType any, VersionType any, GetParamsType any, PutParamsType any] func(r *ResourceType[SourceType, VersionType, GetParamsType, PutParamsType])

func WithStdout[S any, V any, G any, P any](stdout io.Writer) ResourceTypeOption[S, V, G, P] {
	return func(r *ResourceType[S, V, G, P]) {
		r.stdout = stdout
	}
}

func WithStdin[S any, V any, G any, P any](stdin io.Reader) ResourceTypeOption[S, V, G, P] {
	return func(r *ResourceType[S, V, G, P]) {
		r.stdin = stdin
	}
}

func NewResourceType[S any, V any, G any, P any](impl any, options ...ResourceTypeOption[S, V, G, P]) ResourceType[S, V, G, P] {
	r := ResourceType[S, V, G, P]{
		stdout: os.Stdout,
		stdin:  os.Stdin,
	}

	if c, ok := impl.(Checker[S, V]); ok {
		r.check = c
	}

	if i, ok := impl.(Getter[S, V, G]); ok {
		r.in = i
	}

	if o, ok := impl.(Putter[S, V, P]); ok {
		r.out = o
	}

	for _, opt := range options {
		opt(&r)
	}

	return r
}

func (r ResourceType[SourceType, VersionType, GetParamsType, PutParamsType]) Run(args ...string) error {
	if len(args) == 0 {
		args = append([]string{}, os.Args...)
	}

	binPath, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("could not determine path to binary: %w", err)
	}

	switch binPath {
	case "/opt/resource/check":
		if r.check != nil {
			var req CheckRequest[SourceType, VersionType]
			decoder := json.NewDecoder(r.stdin)
			decoder.DisallowUnknownFields()
			if err = decoder.Decode(&req); err != nil {
				return fmt.Errorf("could not decode source: %w", err)
			}

			versions, err := r.check.Check(req)
			if err != nil {
				return fmt.Errorf("check failed: %w", err)
			}

			if err = json.NewEncoder(r.stdout).Encode(versions); err != nil {
				return fmt.Errorf("could not output versions: %w", err)
			}

			return nil
		}

		return errors.New("/opt/resource/check not implemented")

	case "/opt/resource/in":
		if r.in != nil {
			var req GetRequest[SourceType, VersionType, GetParamsType]
			decoder := json.NewDecoder(r.stdin)
			decoder.DisallowUnknownFields()
			if err = decoder.Decode(&req); err != nil {
				return fmt.Errorf("could not decode source: %w", err)
			}

			baseDir := args[1]
			response, err := r.in.Get(baseDir, req)

			if err != nil {
				return fmt.Errorf("get failed: %w", err)
			}

			if err = json.NewEncoder(r.stdout).Encode(response); err != nil {
				return fmt.Errorf("could not output get response: %w", err)
			}

			return nil
		}

		return errors.New("/opt/resource/in not implemented")

	case "/opt/resource/out":
		if r.out != nil {
			var req PutRequest[SourceType, PutParamsType]
			decoder := json.NewDecoder(r.stdin)
			decoder.DisallowUnknownFields()
			if err = decoder.Decode(&req); err != nil {
				return fmt.Errorf("could not decode source: %w", err)
			}

			baseDir := args[1]
			response, err := r.out.Put(baseDir, req)

			if err != nil {
				return fmt.Errorf("get failed: %w", err)
			}

			if err = json.NewEncoder(r.stdout).Encode(response); err != nil {
				return fmt.Errorf("could not output get response: %w", err)
			}

			return nil
		}

		return errors.New("/opt/resource/out not implemented")

	default:
		if stat, _ := os.Stat(binPath); stat == nil || (stat.Mode()&fs.ModeSymlink == fs.ModeSymlink) {
			return fmt.Errorf("%s must not be a symbolic link", binPath)
		}

		if len(args) < 2 || args[1] != "install" {
			return fmt.Errorf("unrecogized arguments %v, only 'install' is allowed", args)
		}

		if err := os.MkdirAll("/opt/resource", 0o777); err != nil {
			return fmt.Errorf("could not ensure /opt/resource exists: %w", err)
		}

		if r.check != nil {
			if err = os.Symlink(binPath, "/opt/resource/check"); err != nil {
				return fmt.Errorf("could not create /opt/resource/check link: %w", err)
			}
		}

		if r.in != nil {
			if err = os.Symlink(binPath, "/opt/resource/in"); err != nil {
				return fmt.Errorf("could not create /opt/resource/in link: %w", err)
			}
		}

		if r.out != nil {
			if err = os.Symlink(binPath, "/opt/resource/out"); err != nil {
				return fmt.Errorf("could not create /opt/resource/out link: %w", err)
			}
		}

		return nil
	}
}
