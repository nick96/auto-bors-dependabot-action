FROM golang:1.15
LABEL org.opencontainers.image.source https://github.com/nick96/auto-bors-dependabot-action

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go install

ENTRYPOINT [ "auto-bors-dependabot-action" ]