package project

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var (
	ErrNotFound  = errors.New("project not found")
	ErrInactive  = errors.New("project inactive")
	ErrDuplicate = errors.New("project already exists")
	ErrInvalidID = errors.New("invalid project id")
)

var validID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

func ValidateID(id string) error {
	if !validID.MatchString(id) {
		return ErrInvalidID
	}
	return nil
}

type Project struct {
	ID     string
	Active bool
}

func (s *Store) Create(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ValidateID(id); err != nil {
		return err
	}
	if _, ok := s.projects[id]; ok {
		return ErrDuplicate
	}
	s.projects[id] = Project{ID: id, Active: true}
	if s.db != nil {
		_, err := s.db.Exec(`INSERT INTO projects (id, active) VALUES (?, 1)`, id)
		if err != nil {
			delete(s.projects, id)
			if strings.Contains(err.Error(), "UNIQUE") {
				return ErrDuplicate
			}
			return err
		}
		return nil
	}
	if err := s.persistLocked(); err != nil {
		delete(s.projects, id)
		return err
	}
	return nil
}

type Store struct {
	mu       sync.RWMutex
	projects map[string]Project
	path     string
	db       *sql.DB
}

func NewStore(projects ...Project) *Store {
	s := &Store{projects: make(map[string]Project, len(projects))}
	for _, p := range projects {
		if ValidateID(p.ID) != nil {
			continue
		}
		s.projects[p.ID] = p
	}
	return s
}

func NewPersistentStore(path string) (*Store, error) {
	store := NewStore()
	store.path = path
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, err
	}
	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, err
	}
	for _, item := range projects {
		if ValidateID(item.ID) == nil {
			store.projects[item.ID] = item
		}
	}
	return store, nil
}

func (s *Store) Resolve(id string) (Project, error) {
	if err := ValidateID(id); err != nil {
		return Project{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db != nil {
		var active bool
		err := s.db.QueryRow(`SELECT active FROM projects WHERE id = ?`, id).Scan(&active)
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, ErrNotFound
		}
		if err != nil {
			return Project{}, err
		}
		if !active {
			return Project{}, ErrInactive
		}
		return Project{ID: id, Active: true}, nil
	}
	p, ok := s.projects[id]
	if !ok {
		return Project{}, ErrNotFound
	}
	if !p.Active {
		return Project{}, ErrInactive
	}
	return p, nil
}

func (s *Store) List() []Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db != nil {
		rows, err := s.db.Query(`SELECT id, active FROM projects ORDER BY id`)
		if err != nil {
			return nil
		}
		defer rows.Close()
		projects := make([]Project, 0)
		for rows.Next() {
			var item Project
			if rows.Scan(&item.ID, &item.Active) == nil {
				projects = append(projects, item)
			}
		}
		return projects
	}
	projects := make([]Project, 0, len(s.projects))
	for _, project := range s.projects {
		projects = append(projects, project)
	}
	sort.Slice(projects, func(left, right int) bool { return projects[left].ID < projects[right].ID })
	return projects
}

func (s *Store) SetActive(id string, active bool) error {
	if err := ValidateID(id); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		result, err := s.db.Exec(`UPDATE projects SET active = ? WHERE id = ?`, active, id)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if count == 0 {
			return ErrNotFound
		}
		return nil
	}
	project, ok := s.projects[id]
	if !ok {
		return ErrNotFound
	}
	previous := project
	project.Active = active
	s.projects[id] = project
	if err := s.persistLocked(); err != nil {
		// Never grant access after a failed durable activation. A failed
		// deactivation remains denied in memory until storage is repaired.
		if active {
			s.projects[id] = previous
		}
		return err
	}
	return nil
}

func (s *Store) persistLocked() error {
	if s.path == "" {
		return nil
	}
	projects := make([]Project, 0, len(s.projects))
	for _, item := range s.projects {
		projects = append(projects, item)
	}
	sort.Slice(projects, func(left, right int) bool { return projects[left].ID < projects[right].ID })
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	temporaryPath := s.path + ".tmp"
	if err := os.WriteFile(temporaryPath, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(temporaryPath, s.path)
}
