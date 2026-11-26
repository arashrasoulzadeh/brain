# Versioned protocol fixtures

The JSON fixtures in `protocol/fixtures/` use `protocol: 1` and are transport
neutral. Clients must preserve the project, room, session, revision, and
sequence fields exactly; unknown fields may be ignored, but missing required
fields are errors. The Go test and Dart consumer validate the published shape.

Clients must reject an unsupported protocol version, missing project identity,
missing room/session identity, invalid revisions, or malformed player input.
They should preserve the original error category and avoid retrying malformed
requests. A reconnect may replay events starting at `lastEventSequence + 1`.

Project IDs match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$` over the entire input.
Do not trim, normalize, or replace invalid characters. Clients reject malformed
IDs before transport; the host remains responsible for authorization.
