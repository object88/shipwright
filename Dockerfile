FROM go:1.16 AS builder

WORKDIR /go/src/github.com/object88/shipwright

COPY . .

RUN build.sh