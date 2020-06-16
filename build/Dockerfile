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
COPY --from=builder /app/durak /app/
RUN mkdir -p /var/log/durak && chmod 0777 /var/log/durak
EXPOSE 8007
ENTRYPOINT ["/app/durak"]
