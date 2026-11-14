import 'package:brain_flutter_client/secure_token_storage.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('storage adapter writes, reads, and clears tokens', () async {
    final storage = MemoryTokenStorage();
    await storage.write('token');
    expect(await storage.read(), 'token');
    await storage.clear();
    expect(await storage.read(), isNull);
  });
}
