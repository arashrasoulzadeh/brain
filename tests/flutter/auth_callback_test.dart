import 'package:brain_flutter_client/auth_callback.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('parses the verified native browser callback', () {
    final callback = parseAuthCallback(
      Uri.parse('https://app.example.test/auth/callback?code=abc&state=xyz'),
    );
    expect(callback.code, 'abc');
    expect(callback.state, 'xyz');
  });

  test('rejects insecure or incomplete callbacks', () {
    expect(
      () => parseAuthCallback(Uri.parse('http://app.example.test/auth/callback?code=abc&state=xyz')),
      throwsFormatException,
    );
    expect(
      () => parseAuthCallback(Uri.parse('https://app.example.test/auth/callback?code=abc')),
      throwsFormatException,
    );
  });
}
