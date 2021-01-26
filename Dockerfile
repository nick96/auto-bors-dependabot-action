FROM golang:1.15

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go install

ENTRYPOINT [ "auto-bors-dependabot-action" ]