bool isSupportedProtocolFixture(Map<String, dynamic> fixture, String name) {
  return fixture['protocol'] == 1 && fixture['fixture'] == name && fixture['projectId'] is String;
}
