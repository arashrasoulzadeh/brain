import 'dart:convert';
import 'dart:io';

import 'package:brain_flutter_client/protocol_fixtures.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('Dart consumer validates the released protocol fixtures', () {
    for (final name in [
      'login',
      'room-admission',
      'assignment',
      'player-input',
      'authorized-snapshot',
      'reconnect',
    ]) {
      final file = File('../../protocol/fixtures/$name.json');
      final fixture = jsonDecode(file.readAsStringSync()) as Map<String, dynamic>;
      expect(isSupportedProtocolFixture(fixture, name), isTrue, reason: name);
    }
  });
}
