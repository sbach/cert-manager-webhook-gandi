FROM golang:1.22.1-alpine3.19 as build_dependencies

RUN apk add --no-cache git

WORKDIR /usr/src/webhook

COPY go.mod .
COPY go.sum .

RUN go mod download


FROM build_dependencies AS build

COPY . .

RUN CGO_ENABLED=0 \
        go build \
            -o webhook \
            -ldflags '-w -extldflags "-static"' \
            .


FROM alpine:3.19

RUN apk add --no-cache ca-certificates

COPY --from=build /usr/src/webhook /usr/local/bin/webhook

ENTRYPOINT ["webhook"]
