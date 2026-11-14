import 'package:brain_flutter_client/web_cookie.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('creates a first-party Flutter Web cookie policy', () {
    expect(
      firstPartyWebCookie('token', 3600),
      '__Host-brain_session=token; Max-Age=3600; Path=/; Secure; SameSite=Lax',
    );
  });
}
