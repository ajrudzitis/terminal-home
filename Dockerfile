# syntax=docker/dockerfile:1

FROM alpine:3.14


# Copy the binary from the build stage
COPY terminal-home /terminal-home

ENV TERM xterm-256color


# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/reference/dockerfile/#expose
EXPOSE 2222

# Run
CMD ["/terminal-home"]