package project

import (
	"errors"
	"regexp"
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
	return nil
}

type Store struct {
	mu       sync.RWMutex
	projects map[string]Project
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

func (s *Store) Resolve(id string) (Project, error) {
	if err := ValidateID(id); err != nil {
		return Project{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.projects[id]
	if !ok {
		return Project{}, ErrNotFound
	}
	if !p.Active {
		return Project{}, ErrInactive
	}
	return p, nil
}
