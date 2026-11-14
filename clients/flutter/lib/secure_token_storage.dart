abstract interface class SecureTokenStorage {
  Future<void> write(String token);

  Future<String?> read();

  Future<void> clear();
}

class MemoryTokenStorage implements SecureTokenStorage {
  String? _token;

  @override
  Future<void> write(String token) async {
    if (token.isEmpty) throw ArgumentError('token required');
    _token = token;
  }

  @override
  Future<String?> read() async => _token;

  @override
  Future<void> clear() async {
    _token = null;
  }
}
