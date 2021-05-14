# use latest 1.15 stable release
# this is in sync with go.mod
FROM golang:1.15

# labels for github actions UI
LABEL "com.github.actions.name"="go-tools"
LABEL "com.github.actions.description"="go tool runner"
LABEL "com.github.actions.icon"="command"
LABEL "com.github.actions.color"="blue"

# usage: in GitHub Actions, use params: "test ./..."
COPY entrypoint /entrypoint
ENTRYPOINT ["/entrypoint"]
