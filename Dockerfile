FROM golang:1.17-alpine3.15 AS build

COPY . /fs-store
WORKDIR /fs-store

# Update/upgrade image (alpine)
RUN apk update --no-cache
RUN apk upgrade --no-cache

RUN go build .


FROM alpine:3.15

COPY --from=build /fs-store/fs-store /usr/bin/fs-store
RUN chmod +x /usr/bin/fs-store

RUN apk update --no-cache
RUN apk upgrade --no-cache

RUN mkdir -p /opt/fs-store/data

## add and change user
RUN adduser 999 -D -S
USER 999

ENTRYPOINT ["fs-store"]
CMD ["server", "--host", "0.0.0.0", "--port", "8080", "--data-dir", "/opt/fs-store/data"]