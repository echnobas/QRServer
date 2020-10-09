FROM golang:latest
WORKDIR /src
COPY . .
ENTRYPOINT ["go", "run", "main.go"]