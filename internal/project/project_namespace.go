package project

import (
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}$`)

func ValidateSlug(slug string) error {
	if slug == "_system" || !slugPattern.MatchString(slug) {
		return ErrInvalidID
	}
	return nil
}

func normalizedProject(id, name, slug string, settings map[string]any) (Project, error) {
	if err := ValidateID(id); err != nil {
		return Project{}, err
	}
	if slug == "" {
		slug = id
	}
	if err := ValidateSlug(slug); err != nil {
		return Project{}, err
	}
	if name == "" {
		name = id
	}
	if strings.TrimSpace(name) != name || len(name) > 128 || strings.ContainsAny(name, "\x00\r\n") {
		return Project{}, ErrInvalidID
	}
	if settings == nil {
		settings = map[string]any{}
	}
	return Project{ID: id, Active: true, Name: name, Slug: slug, Settings: settings}, nil
}

func (s *Store) CreateWithDetails(id, name, slug string, settings map[string]any) error {
	p, err := normalizedProject(id, name, slug, settings)
	if err != nil {
		return err
	}
	data, err := json.Marshal(p.Settings)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		var existing string
		err = s.db.QueryRow(`SELECT id FROM projects WHERE id IN (?,?) OR slug IN (?,?) LIMIT 1`, id, p.Slug, id, p.Slug).Scan(&existing)
		if err == nil {
			return ErrDuplicate
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		_, err = s.db.Exec(`INSERT INTO projects(id,active,name,slug,settings) VALUES (?,1,?,?,?)`, id, p.Name, p.Slug, string(data))
		if err != nil && strings.Contains(err.Error(), "UNIQUE") {
			return ErrDuplicate
		}
		return err
	}
	for _, existing := range s.projects {
		if existing.ID == id || existing.ID == p.Slug || existing.Slug == id || existing.Slug == p.Slug {
			return ErrDuplicate
		}
	}
	s.projects[id] = p
	if err := s.persistLocked(); err != nil {
		delete(s.projects, id)
		return err
	}
	return nil
}

func (s *Store) ResolveNamespace(namespace string) (Project, error) {
	if namespace == "_system" || (ValidateID(namespace) != nil && ValidateSlug(namespace) != nil) {
		return Project{}, ErrInvalidID
	}
	if s.db != nil {
		var id string
		err := s.db.QueryRow(`SELECT id FROM projects WHERE id=? OR slug=?`, namespace, namespace).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, ErrNotFound
		}
		if err != nil {
			return Project{}, err
		}
		return s.Resolve(id)
	}
	s.mu.RLock()
	var id string
	for _, p := range s.projects {
		if p.ID == namespace || p.Slug == namespace {
			id = p.ID
			break
		}
	}
	s.mu.RUnlock()
	if id == "" {
		return Project{}, ErrNotFound
	}
	return s.Resolve(id)
}

func (s *Store) Configure(id, name, slug string, settings map[string]any) error {
	if err := ValidateID(id); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var current Project
	if s.db != nil {
		var settingsJSON string
		err := s.db.QueryRow(`SELECT id,active,name,slug,settings FROM projects WHERE id=?`, id).
			Scan(&current.ID, &current.Active, &current.Name, &current.Slug, &settingsJSON)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(settingsJSON), &current.Settings); err != nil {
			return err
		}
	} else {
		var exists bool
		current, exists = s.projects[id]
		if !exists {
			return ErrNotFound
		}
	}
	if name == "" {
		name = current.Name
	}
	if slug == "" {
		slug = current.Slug
	}
	p, err := normalizedProject(id, name, slug, settings)
	if err != nil {
		return err
	}
	p.Active = current.Active
	if settings == nil {
		p.Settings = current.Settings
	}
	data, err := json.Marshal(p.Settings)
	if err != nil {
		return err
	}
	if s.db != nil {
		var collision string
		err = s.db.QueryRow(`SELECT id FROM projects WHERE id<>? AND (id IN (?,?) OR slug IN (?,?)) LIMIT 1`, id, id, slug, id, slug).Scan(&collision)
		if err == nil {
			return ErrDuplicate
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		_, err = s.db.Exec(`UPDATE projects SET name=?,slug=?,settings=? WHERE id=?`, p.Name, p.Slug, string(data), id)
		if err != nil && strings.Contains(err.Error(), "UNIQUE") {
			return ErrDuplicate
		}
		return err
	}
	for otherID, existing := range s.projects {
		if otherID != id && (existing.ID == slug || existing.Slug == slug) {
			return ErrDuplicate
		}
	}
	s.projects[id] = p
	if err := s.persistLocked(); err != nil {
		s.projects[id] = current
		return err
	}
	return nil
}
