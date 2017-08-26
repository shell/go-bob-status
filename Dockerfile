FROM scratch
ADD go-bob-status /

ARG github_token
ARG jenkins_user
ARG jenkins_password
CMD ["/go-bob-status", "-u", "$jenkins_user", "-p", "$jenkins_password", "-t", "$github_token"]
