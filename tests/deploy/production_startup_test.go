package deploy_test

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProductionBinaryRejectsUnsafeStartupConfiguration(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "brain")
	build := exec.Command("go", "build", "-buildvcs=false", "-o", binary, "./cmd/brain")
	build.Dir = "../.."
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build host: %v\n%s", err, output)
	}
	for _, tc := range []struct {
		name string
		env  []string
		want string
	}{
		{name: "missing operator", want: "OPERATOR_USERNAME and OPERATOR_SECRET are required"},
		{name: "missing storage", env: []string{"OPERATOR_USERNAME=ci", "OPERATOR_SECRET=fixture"}, want: "PROJECTS_FILE must name an explicit durable SQLite file"},
		{name: "memory storage", env: []string{"OPERATOR_USERNAME=ci", "OPERATOR_SECRET=fixture", "PROJECTS_FILE=:memory:"}, want: "PROJECTS_FILE must name an explicit durable SQLite file"},
		{name: "unsupported database", env: []string{"OPERATOR_USERNAME=ci", "OPERATOR_SECRET=fixture", "RUNTIME_DSN=postgres://fixture", "PROJECTS_FILE=brain.sqlite"}, want: "refusing implicit SQLite fallback"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, binary)
			command.Env = append([]string{"HTTP_ADDRESS=127.0.0.1:0"}, tc.env...)
			output, err := command.CombinedOutput()
			if err == nil || ctx.Err() != nil || !strings.Contains(string(output), tc.want) {
				t.Fatalf("unsafe startup not rejected as expected: %v, %s", err, output)
			}
		})
	}
}
