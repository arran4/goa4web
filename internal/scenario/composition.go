package scenario

import (
	"errors"
	"fmt"
	"strings"
)

// Compose joins scenarios into one explicitly ordered scenario. Symbolic references
// are global to the composition, so later scenarios may refer to earlier ones.
func Compose(scenarios ...*Scenario) (*Scenario, error) {
	if len(scenarios) == 0 {
		return nil, errors.New("at least one scenario is required")
	}

	combined := &Scenario{Meta: Meta{Format: FormatV1}}
	names := make([]string, 0, len(scenarios))
	for idx, sc := range scenarios {
		if sc == nil {
			return nil, fmt.Errorf("scenario[%d]: nil scenario", idx)
		}
		if err := validateMeta(sc); err != nil {
			return nil, fmt.Errorf("scenario[%d]: %w", idx, err)
		}
		names = append(names, sc.Meta.Name)
		for _, evt := range sc.Events {
			if evt == nil {
				return nil, fmt.Errorf("scenario %q: nil event", sc.Meta.Name)
			}
			copy := *evt
			copy.File = sc.Meta.Name + "/" + evt.File
			if copy.FS == nil {
				copy.FS = sc.FS
			}
			combined.Events = append(combined.Events, &copy)
		}
	}
	combined.Meta.Name = strings.Join(names, " + ")
	return combined, nil
}
