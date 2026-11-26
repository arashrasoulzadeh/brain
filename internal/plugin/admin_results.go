package plugin

import "strings"

// AdminResultSpec selects safe presentation components, never executable code.
type AdminResultSpec struct {
	Type       string            `yaml:"type" json:"type"`
	Title      string            `yaml:"title,omitempty" json:"title,omitempty"`
	Path       string            `yaml:"path,omitempty" json:"path,omitempty"`
	Columns    []string          `yaml:"columns,omitempty" json:"columns,omitempty"`
	PageSize   int               `yaml:"page_size,omitempty" json:"page_size,omitempty"`
	LabelField string            `yaml:"label_field,omitempty" json:"label_field,omitempty"`
	ValueField string            `yaml:"value_field,omitempty" json:"value_field,omitempty"`
	Children   []AdminResultSpec `yaml:"children,omitempty" json:"children,omitempty"`
}

func (s AdminResultSpec) valid(depth int) bool {
	if depth > 3 || len(s.Title) > 120 || len(s.Path) > 256 || len(s.Columns) > 32 || s.PageSize < 0 || s.PageSize > 100 || len(s.Children) > 8 {
		return false
	}
	if s.Path != "" {
		for _, key := range strings.Split(s.Path, ".") {
			if !validResponseKey(key) {
				return false
			}
		}
	}
	columns := map[string]bool{}
	for _, column := range s.Columns {
		if !validResponseKey(column) || columns[column] {
			return false
		}
		columns[column] = true
	}
	switch s.Type {
	case "table", "detail", "metric", "json", "code", "logs":
		return len(s.Children) == 0
	case "chart":
		return len(s.Children) == 0 && validResponseKey(s.LabelField) && validResponseKey(s.ValueField)
	case "grid", "tabs":
		if len(s.Children) == 0 {
			return false
		}
		for _, child := range s.Children {
			if !child.valid(depth + 1) {
				return false
			}
		}
		return true
	}
	return false
}
