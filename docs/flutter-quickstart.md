# Flutter quickstart

Run the client package with the Flutter SDK:

```sh
cd clients/flutter
PUB_HOSTED_URL=https://pub.flutter-io.cn flutter pub get
PUB_HOSTED_URL=https://pub.flutter-io.cn flutter test ../../tests/flutter
```

Use `NativeTransport` for a native application and `WebTransport` for Flutter
Web. Both adapters require a project ID and forward typed operations through
the host boundary. Store refresh credentials behind `SecureTokenStorage`; do
not place them in widget state or logs.
