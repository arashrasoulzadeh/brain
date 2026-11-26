/// Matches the host project identity grammar without trimming or normalization.
void validateProjectId(String projectId) {
  if (projectId.isEmpty) throw ArgumentError('project required');
  final match = RegExp(r'^[a-zA-Z0-9][a-zA-Z0-9_-]*').firstMatch(projectId);
  if (match == null || match.end != projectId.length) {
    throw ArgumentError('invalid project ID');
  }
}
