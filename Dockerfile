FROM golang:1.16.4-alpine3.13 as build
WORKDIR /app
ADD . /app
RUN apk add --no-cache build-base sqlite-dev && cd /app && go build

FROM alpine:3.13.4 as production
COPY --from=build /app/newshub-server .
CMD ["./newshub-server"]