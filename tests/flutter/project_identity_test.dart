import 'package:brain_flutter_client/native_transport.dart';
import 'package:brain_flutter_client/web_transport.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('both transports enforce the host project identity grammar', () {
    for (final id in ['', ' ', '../dev', 'dev/project', '-dev', '_dev', 'dev\n', 'dev space', 'é']) {
      expect(() => NativeTransport((_, __, ___) async => null, id), throwsArgumentError);
      expect(() => WebTransport((_, __, ___, ____) async => null, id), throwsArgumentError);
    }
    for (final id in ['dev', 'A', '9', 'dev-01_test']) {
      expect(NativeTransport((_, __, ___) async => null, id).projectId, id);
      expect(WebTransport((_, __, ___, ____) async => null, id).projectId, id);
    }
  });
}
