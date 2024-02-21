FROM golang:1.22 as builder

WORKDIR /app
COPY . /app

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o rinha ./cmd/main.go

FROM scratch

COPY go.mod go.sum ./
WORKDIR /app
COPY --from=builder /app/rinha ./

CMD [ "./rinha" ]