FROM golang:1.19-alpine as build


WORKDIR /work

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o tenant-controller main.go

FROM alpine:latest

COPY --from=build /work/tenant-controller .

EXPOSE 8080

ENTRYPOINT ["./tenant-controller"]