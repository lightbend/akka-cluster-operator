# use latest 1.x stable release
FROM golang:1

# labels for github actions UI
LABEL "com.github.actions.name"="go-tools"
LABEL "com.github.actions.description"="go tool runner"
LABEL "com.github.actions.icon"="command"
LABEL "com.github.actions.color"="blue"

# usage: in GitHub Actions, use params: "test ./..."
COPY entrypoint /entrypoint
ENTRYPOINT ["/entrypoint"]
