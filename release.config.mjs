export default {
  branches: ["master"],
  plugins: [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    [
      "@semantic-release/exec",
      {
        publishCmd:
          "docker buildx imagetools create $IMAGE:$GITHUB_SHA --tag $IMAGE:${nextRelease.version}",
      },
    ],
    "@semantic-release/github",
  ],
};
