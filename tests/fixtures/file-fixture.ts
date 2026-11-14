export type FileFixture = {
  projectId: string;
  bucket: string;
  objectKey: string;
  contentType: "text/plain";
  bytes: string;
};

export function createFileFixture(projectId = "dev"): FileFixture {
  if (!projectId.trim()) throw new Error("fixture project required");
  return {
    projectId,
    bucket: "fixture-files",
    objectKey: "profiles/fixture.txt",
    contentType: "text/plain",
    bytes: "fixture file content",
  };
}
