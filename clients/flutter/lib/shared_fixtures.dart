class SharedFixtureConsumer {
  const SharedFixtureConsumer();

  bool acceptsSignIn(Map<String, Object?> fixture) {
    return fixture['projectId'] is String &&
        (fixture['projectId']! as String).isNotEmpty &&
        fixture['username'] is String &&
        fixture['password'] is String;
  }
}
