# build stage
FROM golang as builder

WORKDIR /app

COPY ./go.mod .
COPY ./go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# final stage
FROM busybox
WORKDIR /app
RUN mkdir -p /var/log/durak && chmod 0777 /var/log/durak
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/durak /app/
COPY ./html/ /app/html/
COPY ./version /app/version
EXPOSE 8007
ENTRYPOINT ["/app/durak", "--addr=:8007", "--env=production"]
