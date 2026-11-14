# Versioned protocol fixtures

The JSON fixtures in `protocol/fixtures/` use `protocol: 1` and are transport
neutral. Clients must preserve the project, room, session, revision, and
sequence fields exactly; unknown fields may be ignored, but missing required
fields are errors. The Go test and Dart consumer validate the published shape.
