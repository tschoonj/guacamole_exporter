FROM golang:1.14.3-alpine AS build
WORKDIR /src
COPY . .
RUN go build
 
FROM alpine
RUN apk --no-cache add ca-certificates tini
WORKDIR /
COPY --from=build /src/guacamole_exporter .
ENTRYPOINT ["/sbin/tini", "--", "/guacamole_exporter"]

EXPOSE 9623