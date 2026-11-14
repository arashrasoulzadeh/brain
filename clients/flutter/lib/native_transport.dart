typedef NativeRequest = Future<Object?> Function(
  String operation,
  String projectId,
  Object? payload,
);

class NativeTransport {
  NativeTransport(this._request, this.projectId) {
    if (projectId.isEmpty) throw ArgumentError('project required');
  }

  final NativeRequest _request;
  final String projectId;

  Future<Object?> send(String operation, Object? payload) {
    if (operation.isEmpty) throw ArgumentError('operation required');
    return _request(operation, projectId, payload);
  }
}
