#docker build -t go-rest-balance .
#docker run -dit --name go-rest-balance -p 3000:3000 go-rest-balance

FROM golang:1.21 As builder

WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go build -o go-rest-balance -ldflags '-linkmode external -w -extldflags "-static"'

FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/go-rest-balance .

CMD ["/app/go-rest-balance"]