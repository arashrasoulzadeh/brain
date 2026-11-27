# Flutter quickstart

Install the Flutter SDK and run:

```sh
cd clients/flutter
flutter pub get
flutter test ../../tests/flutter
```

`NativeTransport` wraps a request callback receiving operation, project ID and payload; `WebTransport` additionally supplies JSON and client headers. Both validate the project identifier and reject empty operations. Connect either to your trusted app transport, not a nonexistent generic Brain HTTP route.

```dart
final transport = NativeTransport(myTrustedRequest, 'dev');
final auth = FlutterAuthClient(transport.send);
final session = await auth.signIn('player', 'example-password');
final account = await auth.session(session['token'] as String);
```

Import the transport and auth client from the package's lib directory. `myTrustedRequest` is your application implementation; see [HTTP API](http-api.md) for mounted endpoints. `FlutterAuthClient` validates token/username response shapes. `SecureTokenStorage` is an interface; the included `MemoryTokenStorage` is volatile and not secure persistent storage. Supply an OS-backed implementation for production. Do not log tokens. Web-cookie and callback validators are utilities, not an integrated browser auth server.

The Dart tests and coverage gate are in the repository's [quality workflow](../.github/workflows/coverage.yml).
