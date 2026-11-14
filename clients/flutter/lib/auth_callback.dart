class AuthCallback {
  const AuthCallback({required this.code, required this.state});

  final String code;
  final String state;
}

AuthCallback parseAuthCallback(Uri uri) {
  if (uri.scheme != 'https' || uri.path != '/auth/callback') {
    throw FormatException('invalid auth callback');
  }
  final code = uri.queryParameters['code'];
  final state = uri.queryParameters['state'];
  if (code == null || code.isEmpty || state == null || state.isEmpty) {
    throw FormatException('invalid auth callback');
  }
  return AuthCallback(code: code, state: state);
}
