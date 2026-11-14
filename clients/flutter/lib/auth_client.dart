import 'native_transport.dart';

class FlutterAuthClient {
  FlutterAuthClient(this._transport);

  final Future<Object?> Function(String operation, Object? payload) _transport;

  Future<Map<String, Object?>> signIn(String username, String password) async {
    final result = await _transport('auth.signin', {'username': username, 'password': password});
    if (result is! Map<String, Object?> || result['token'] is! String) {
      throw StateError('invalid sign-in response');
    }
    return result;
  }

  Future<Map<String, Object?>> session(String token) async {
    final result = await _transport('auth.session', {'token': token});
    if (result is! Map<String, Object?> || result['username'] is! String) {
      throw StateError('invalid session response');
    }
    return result;
  }
}
