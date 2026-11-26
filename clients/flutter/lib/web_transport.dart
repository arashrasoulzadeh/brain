import 'project_identity.dart';

typedef WebRequest = Future<Object?> Function(
  String operation,
  String projectId,
  Object? payload,
  Map<String, String> headers,
);

class WebTransport {
  WebTransport(this._request, this.projectId) {
    validateProjectId(projectId);
  }

  final WebRequest _request;
  final String projectId;

  Future<Object?> send(String operation, Object? payload) {
    if (operation.isEmpty) throw ArgumentError('operation required');
    return _request(operation, projectId, payload, const {
      'content-type': 'application/json',
      'x-client': 'brain-flutter-web',
    });
  }
}
