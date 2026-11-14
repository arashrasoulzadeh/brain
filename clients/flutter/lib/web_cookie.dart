String firstPartyWebCookie(String token, int maxAgeSeconds) {
  if (token.isEmpty || maxAgeSeconds <= 0 || token.contains(';') || token.contains(' ')) {
    throw ArgumentError('invalid web cookie');
  }
  return '__Host-brain_session=$token; Max-Age=$maxAgeSeconds; Path=/; Secure; SameSite=Lax';
}
