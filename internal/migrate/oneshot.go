package migrate

import "context"

// RunOneShot executes the supplied release migrations once for one deployment
// process. The mark callback persists successful versions for future releases.
func RunOneShot(ctx context.Context, migrations []Migration, mark func(context.Context, int) error) error {
	return NewRunner(migrations, mark).Run(ctx)
}
