package project

import "context"

type contextKey struct{}

func WithProject(ctx context.Context, p Project) context.Context {
	return context.WithValue(ctx, contextKey{}, p)
}

func FromContext(ctx context.Context) (Project, bool) {
	p, ok := ctx.Value(contextKey{}).(Project)
	return p, ok
}
