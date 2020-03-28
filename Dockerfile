FROM golang:1.14.1-buster AS builder
WORKDIR /root
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ssmloader .

FROM gcr.io/distroless/static-debian10
COPY --from=builder /root/ssmloader ./
ENTRYPOINT ["/ssmloader"]
