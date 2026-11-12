package project

import "errors"

var (
	ErrNotFound  = errors.New("project not found")
	ErrInactive  = errors.New("project inactive")
	ErrDuplicate = errors.New("project already exists")
)

type Project struct {
	ID     string
	Active bool
}

func (s *Store) Create(id string) error {
	if id == "" {
		return ErrNotFound
	}
	if _, ok := s.projects[id]; ok {
		return ErrDuplicate
	}
	s.projects[id] = Project{ID: id, Active: true}
	return nil
}

type Store struct {
	projects map[string]Project
}

func NewStore(projects ...Project) *Store {
	s := &Store{projects: make(map[string]Project, len(projects))}
	for _, p := range projects {
		s.projects[p.ID] = p
	}
	return s
}

func (s *Store) Resolve(id string) (Project, error) {
	p, ok := s.projects[id]
	if !ok {
		return Project{}, ErrNotFound
	}
	if !p.Active {
		return Project{}, ErrInactive
	}
	return p, nil
}
