package scenario

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
	"time"

	"golang.org/x/tools/txtar"
)

// ErrInvalidFormat indicates that scenario.meta did not contain the expected Format header.
type ErrInvalidFormat struct {
	Format string
}

func (e ErrInvalidFormat) Error() string {
	return fmt.Sprintf("invalid scenario format %q, expected %q", e.Format, FormatV1)
}

// ErrMissingMeta indicates that scenario.meta was not found in the TXTAR archive.
var ErrMissingMeta = errors.New("missing required file scenario.meta in scenario")

// Parse parses a scenario from raw TXTAR bytes.
// If fsys is provided, relative asset paths will be resolved and validated against it.
func Parse(data []byte, fsys fs.FS) (*Scenario, error) {
	ar := txtar.Parse(data)
	if ar == nil {
		return nil, errors.New("failed to parse txtar archive")
	}

	var meta *Meta
	var events []*Event

	for _, f := range ar.Files {
		name := strings.TrimSpace(f.Name)
		if name == MetaFile {
			if meta != nil {
				return nil, fmt.Errorf("duplicate %s in scenario archive", MetaFile)
			}
			parsedMeta, err := parseMeta(f.Data)
			if err != nil {
				return nil, fmt.Errorf("parse %s: %w", MetaFile, err)
			}
			meta = parsedMeta
		} else if strings.HasSuffix(name, ".event") {
			evt, err := parseEvent(name, f.Data)
			if err != nil {
				return nil, fmt.Errorf("parse event %s: %w", name, err)
			}
			events = append(events, evt)
		} else {
			return nil, fmt.Errorf("unexpected file %q in scenario archive (expected %s or *.event)", name, MetaFile)
		}
	}

	if meta == nil {
		return nil, ErrMissingMeta
	}

	if meta.Format != FormatV1 {
		return nil, ErrInvalidFormat{Format: meta.Format}
	}

	return &Scenario{
		Meta:   *meta,
		Events: events,
		FS:     fsys,
	}, nil
}

// ParseReader parses a scenario from an io.Reader containing TXTAR data.
func ParseReader(r io.Reader, fsys fs.FS) (*Scenario, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read scenario: %w", err)
	}
	return Parse(data, fsys)
}

// ParseFS reads and parses a scenario TXTAR file from an fs.FS.
// The provided fsys is also used as the filesystem context for adjacent asset lookups
// scoped to the directory containing the scenario file.
func ParseFS(fsys fs.FS, scenarioPath string) (*Scenario, error) {
	cleanPath := path.Clean(scenarioPath)
	data, err := fs.ReadFile(fsys, cleanPath)
	if err != nil {
		return nil, fmt.Errorf("read scenario file %s: %w", scenarioPath, err)
	}

	dir := path.Dir(cleanPath)
	assetFS := fsys
	if dir != "." && dir != "" && dir != "/" {
		sub, err := fs.Sub(fsys, dir)
		if err != nil {
			return nil, fmt.Errorf("create scenario sub-filesystem for %s: %w", dir, err)
		}
		assetFS = sub
	}

	return Parse(data, assetFS)
}

func parseMeta(data []byte) (*Meta, error) {
	headers, _, err := ParseHeadersAndBody(data)
	if err != nil {
		return nil, err
	}

	format := strings.TrimSpace(headers.Get("Format"))
	name := strings.TrimSpace(headers.Get("Name"))
	desc := strings.TrimSpace(headers.Get("Description"))

	return &Meta{
		Format:      format,
		Name:        name,
		Description: desc,
		Headers:     headers,
	}, nil
}

func parseEvent(filename string, data []byte) (*Event, error) {
	headers, body, err := ParseHeadersAndBody(data)
	if err != nil {
		return nil, err
	}

	op := strings.TrimSpace(headers.Get("Op"))
	if op == "" {
		return nil, fmt.Errorf("missing required header 'Op'")
	}

	ref := strings.TrimSpace(headers.Get("Ref"))
	atStr := strings.TrimSpace(headers.Get("At"))
	var at time.Time
	if atStr != "" {
		parsedAt, err := time.Parse(time.RFC3339, atStr)
		if err != nil {
			// Try RFC3339Nano as well
			parsedAt, err = time.Parse(time.RFC3339Nano, atStr)
			if err != nil {
				return nil, fmt.Errorf("invalid timestamp %q in 'At': %w", atStr, err)
			}
		}
		at = parsedAt
	}

	return &Event{
		File:    filename,
		Op:      op,
		Ref:     ref,
		At:      at,
		Headers: headers,
		Body:    body,
	}, nil
}

// ParseHeadersAndBody parses RFC822 / HTTP style headers and an optional body.
// Headers and body are separated by the first empty line.
func ParseHeadersAndBody(data []byte) (Header, string, error) {
	headers := NewHeader()
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var inBody bool
	var bodyLines []string

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if inBody {
			bodyLines = append(bodyLines, line)
			continue
		}

		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" {
			inBody = true
			continue
		}

		// Comment in header section (starts with #) is ignored
		if strings.HasPrefix(strings.TrimSpace(trimmed), "#") {
			continue
		}

		colonIdx := strings.IndexByte(trimmed, ':')
		if colonIdx == -1 {
			return headers, "", fmt.Errorf("line %d: malformed header %q (missing ':')", lineNum, trimmed)
		}

		key := strings.TrimSpace(trimmed[:colonIdx])
		if key == "" {
			return headers, "", fmt.Errorf("line %d: empty header key in %q", lineNum, trimmed)
		}
		val := strings.TrimSpace(trimmed[colonIdx+1:])

		headers.Add(key, val)
	}

	if err := scanner.Err(); err != nil {
		return headers, "", err
	}

	var body string
	if len(bodyLines) > 0 {
		body = strings.Join(bodyLines, "\n")
	}

	return headers, body, nil
}
