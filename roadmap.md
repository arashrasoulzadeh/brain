# Brain roadmap index

Work in this order:

1. [Backend roadmap](roadmap_backend.md)
2. [Admin roadmap](roadmap_admin.md)
3. [Client roadmap](roadmap_client.md)

Every checkbox is a small delivery step. For each checkbox:

- Write a focused failing test.
- Implement the smallest passing change.
- Run focused and affected tests.
- Update `CHANGELOG.md`, `docs/developers.md`, and `docs/admins.md` when the change affects a phase or operator behavior.
- Commit the verified checkbox separately.
- Push the commit and confirm the remote contains it.

Backend work is the current priority. Do not begin admin or client implementation until the backend roadmap is complete and its release gates pass.
\n- [ ] Complete and verify the project-scoped routing migration, including CI, cross-platform clients, security tests, documentation, and a green merge to main (PR #38).\n