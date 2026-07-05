# Uses a pre-built Linux binary from bin/server-linux.
# Build it locally with:
#   GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bin/server-linux ./cmd/server
# Then rebuild the image:
#   docker compose up -d --build server
FROM scratch

COPY bin/server-linux /server
COPY migrations /migrations

EXPOSE 8080

CMD ["/server"]
