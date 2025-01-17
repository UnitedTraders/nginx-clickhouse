# build stage
FROM golang:1.17-alpine AS build-env

WORKDIR /go/src/github.com/UnitedTraders/nginx-clickhouse

ADD . /go/src/github.com/UnitedTraders/nginx-clickhouse

RUN apk update && apk add make g++ git curl
RUN cd /go/src/github.com/UnitedTraders/nginx-clickhouse && go get . 
RUN cd /go/src/github.com/UnitedTraders/nginx-clickhouse && make build

# final stage
FROM scratch

COPY --from=build-env /go/src/github.com/UnitedTraders/nginx-clickhouse/nginx-clickhouse /
CMD [ "/nginx-clickhouse" ]

