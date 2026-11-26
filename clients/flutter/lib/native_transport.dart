import 'project_identity.dart';

typedef NativeRequest = Future<Object?> Function(
  String operation,
  String projectId,
  Object? payload,
);

class NativeTransport {
  NativeTransport(this._request, this.projectId) {
    validateProjectId(projectId);
  }

  final NativeRequest _request;
  final String projectId;

  Future<Object?> send(String operation, Object? payload) {
    if (operation.isEmpty) throw ArgumentError('operation required');
    return _request(operation, projectId, payload);
  }
}
