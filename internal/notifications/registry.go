package notifications

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web/internal/eventbus"
	"golang.org/x/tools/txtar"
)

type NotificationConfig struct {
	EventPattern   string
	RequiredGrants []GrantRequirement
	DefaultRoles   []string
	RequiredTiers  []string
	Templates      map[string]string
}

func ParseTxtarConfig(data []byte) (*NotificationConfig, error) {
	archive := txtar.Parse(data)
	if len(archive.Files) == 0 {
		return nil, fmt.Errorf("no files found in txtar")
	}

	config := &NotificationConfig{
		Templates: make(map[string]string),
	}

	meta := string(bytes.TrimSpace(archive.Comment))
	lines := strings.Split(meta, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "EventPattern":
			config.EventPattern = val
		case "DefaultRoles":
			config.DefaultRoles = strings.Split(val, ",")
			for i := range config.DefaultRoles {
				config.DefaultRoles[i] = strings.TrimSpace(config.DefaultRoles[i])
			}
		case "RequiredTiers":
			config.RequiredTiers = strings.Split(val, ",")
			for i := range config.RequiredTiers {
				config.RequiredTiers[i] = strings.TrimSpace(config.RequiredTiers[i])
			}
		}
	}

	for _, file := range archive.Files {
		config.Templates[file.Name] = strings.TrimSpace(string(file.Data))
	}

	return config, nil
}

type Registry interface {
	Load() error
	ProcessEvent(ctx context.Context, evt eventbus.TaskEvent) error
}

type MemoryRegistry struct {
	notifier *Notifier
	configs  []*NotificationConfig
}

func NewRegistry(notifier *Notifier) *MemoryRegistry {
	return &MemoryRegistry{
		notifier: notifier,
	}
}

func (r *MemoryRegistry) LoadFromFS(fsys fs.FS, dir string) error {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".txtar" {
			continue
		}
		data, err := fs.ReadFile(fsys, filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		cfg, err := ParseTxtarConfig(data)
		if err != nil {
			return err
		}
		r.AddConfig(cfg)
	}
	return nil
}

func (r *MemoryRegistry) Load() error {
	// For MVP, start with an empty set of configurations
	return nil
}

func (r *MemoryRegistry) AddConfig(cfg *NotificationConfig) {
	r.configs = append(r.configs, cfg)
}

func (r *MemoryRegistry) ProcessEvent(ctx context.Context, evt eventbus.TaskEvent) error {
	if r.notifier == nil {
		return nil
	}

	log.Printf("ProcessEvent triggered for path %s", evt.Path)
	return nil
}
