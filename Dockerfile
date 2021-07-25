FROM golang:1.16 AS builder

WORKDIR /go/src/github.com/object88/shipwright

COPY . .

RUN ./build.sh

FROM scratch AS release

USER appuser

CMD ["/usr/local/bin/shipwright", "run", "--verbose"]

# Keep this late to minimize the number of layer changes.
COPY --from=builder "/go/src/github.com/object88/shipwright/bin/shipwright-linux-amd64" "/usr/local/bin/shipwright"
