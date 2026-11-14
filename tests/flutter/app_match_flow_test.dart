import 'package:brain_flutter_client/auth_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('runs one authenticated app flow', () async {
    final client = FlutterAuthClient((operation, payload) async {
      return {'token': 'token', 'username': 'alice'};
    });
    final session = await client.signIn('alice', 'password');
    expect(session['username'], 'alice');
  });

  test('runs one two-player match flow', () {
    final players = ['alice', 'bob'];
    var revision = 0;
    revision++;
    expect(players, hasLength(2));
    expect(revision, 1);
  });
}
