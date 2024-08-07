# FROM debian:buster
FROM golang:1.21 AS thanos-build

ARG TARGETARCH
ARG TARGETOS

RUN mkdir /thanos-events

COPY ./ /thanos-events/

WORKDIR /thanos-events

RUN go mod tidy
RUN env GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o thanos-app-notif ./cmd/app/main.go

FROM  alpine as certs
RUN apk update && apk add ca-certificates

FROM  --platform=${TARGETOS}/${TARGETARCH} busybox:latest
COPY --from=thanos-build /thanos-events/thanos-app-notif /bin/thanos-app-notif
COPY --from=certs /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT [ "thanos-app-notif" ]

CMD ["listener"]