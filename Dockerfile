FROM golang:onbuild
ADD . /go/src/github.com/shell/go-bob-status
RUN go install github.com/shell/go-bob-status
CMD ["/go/bin/go-bob-status", "-t",  "$GITHUB_TOKEN", "-u", "$JENKINS_USER", "-p", "$JENKINS_PASSWORD"]