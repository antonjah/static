FROM golang:1.19-alpine AS build

WORKDIR /go/src

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/static cmd/static/static.go

FROM scratch

COPY --from=build /go/bin/static /bin/static

CMD ["/bin/static"]