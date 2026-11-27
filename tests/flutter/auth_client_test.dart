import 'package:brain_flutter_client/auth_client.dart';
import 'package:brain_flutter_client/native_transport.dart';
import 'package:brain_flutter_client/web_transport.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('completes sign-in and session on native transport', () async {
    final transport = NativeTransport((operation, projectId, payload) async {
      if (operation == 'auth.signin') return {'token': 'token', 'username': 'alice'};
      return {'username': 'alice'};
    }, 'dev');
    final client = FlutterAuthClient((operation, payload) => transport.send(operation, payload));
    final signedIn = await client.signIn('alice', 'password');
    expect(signedIn['token'], 'token');
    expect((await client.session('token'))['username'], 'alice');
  });

  test('completes sign-in on web transport', () async {
    final transport = WebTransport((operation, projectId, payload, headers) async {
      return {'token': 'token', 'username': 'alice'};
    }, 'dev');
    final client = FlutterAuthClient((operation, payload) => transport.send(operation, payload));
    expect((await client.signIn('alice', 'password'))['username'], 'alice');
  });
  test('rejects non-map and malformed sign-in and session responses', () async {
    for (final value in <Object?>[null, 'not a response', <String, Object?>{},
      {'token': 7, 'username': 9}]) {
      final client = FlutterAuthClient((operation, payload) async => value);
      await expectLater(client.signIn('alice', 'password'), throwsStateError);
      await expectLater(client.session('token'), throwsStateError);
    }
  });
}
