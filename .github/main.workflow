workflow "test on PRs" {
  on = "pull_request"
  resolves = ["go-tools"]
}

action "go-tools" {
  uses = "./.github/action/go-tools"
  args = "test ./..."
}
