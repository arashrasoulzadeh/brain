import 'package:brain_flutter_client/web_transport.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('web transport sends JSON requests with browser client headers', () async {
    Map<String, String>? receivedHeaders;
    final transport = WebTransport((operation, projectId, payload, headers) async {
      receivedHeaders = headers;
      return '$projectId:$operation:$payload';
    }, 'dev');
    expect(await transport.send('data.read', {'id': 'one'}), 'dev:data.read:{id: one}');
    expect(receivedHeaders?['content-type'], 'application/json');
  });
}
