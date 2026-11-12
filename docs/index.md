# Brain mBaaS documentation

Brain is a project-isolated Go backend for apps, websites, and games.

## Start here

- [Developer guide](developers.md)
- [Administrator guide](admins.md)
- [Phase options and decisions](phase-options.md)
- [Roadmap](https://github.com/arashrasoulzadeh/brain/blob/main/roadmap.md)
- [Changelog](https://github.com/arashrasoulzadeh/brain/blob/main/CHANGELOG.md)

## Current status

Phase 00 is in progress. Local tests use an in-process host and isolated test settings. Docker is reserved for explicit integration and release checks.

## Design principles

- Project identity is carried through every protected operation.
- Authorization is checked by the host, not trusted to workers or clients.
- Runtime and migration database credentials are separate.
- Each verified slice is tested, committed, and pushed before the next slice.
