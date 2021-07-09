ARG VERSION="1.0.0"

FROM golang:1.16-alpine3.13 as build

RUN apk update && apk add git make curl

WORKDIR /go/src/immune
COPY . .
RUN make download
ENV VERSION=${VERSION}
ENV NAME="immune"

RUN make test
RUN make build

FROM alpine:3.13 as bin
ARG NAME

COPY --from=build /go/src/immune/immune /immune

ENTRYPOINT ["/immune"]
CMD ["-config=/configs/config.json", "-values=/configs/values.json"]