import 'package:brain_flutter_client/shared_fixtures.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('consumes the shared sign-in fixture contract', () {
    const consumer = SharedFixtureConsumer();
    expect(
      consumer.acceptsSignIn({
        'projectId': 'dev',
        'username': 'fixture-user@example.test',
        'password': 'fixture-password-123',
      }),
      isTrue,
    );
    expect(
      consumer.acceptsSignIn({
        'projectId': '',
        'username': 'x',
        'password': 'y',
      }),
      isFalse,
    );
  });
}
