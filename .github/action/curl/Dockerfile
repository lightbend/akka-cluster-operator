FROM alpine:latest

# labels for github actions UI
LABEL "com.github.actions.name"="curl"
LABEL "com.github.actions.description"="curl runner"
LABEL "com.github.actions.icon"="upload"
LABEL "com.github.actions.color"="black"

RUN apk add --update curl

# usage: in GitHub Actions, use params like: -d "repo=github.com/${GITHUB_REPOSITORY}" https://goreportcard.com/checks
ENTRYPOINT ["curl"]
