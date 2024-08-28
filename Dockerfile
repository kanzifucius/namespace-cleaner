FROM golang:alpine AS build-env
ADD . /src
WORKDIR /src

RUN CGO_ENABLED=0 go test  ./...
RUN go build -o goapp

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/goapp /app/
ENTRYPOINT ./goapp