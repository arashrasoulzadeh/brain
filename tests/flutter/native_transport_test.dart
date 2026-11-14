import 'package:brain_flutter_client/native_transport.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('native transport binds every request to its project', () async {
    String? receivedProject;
    final transport = NativeTransport((operation, projectId, payload) async {
      receivedProject = projectId;
      return '$operation:$payload';
    }, 'dev');
    expect(await transport.send('auth.session', 'token'), 'auth.session:token');
    expect(receivedProject, 'dev');
  });
}
