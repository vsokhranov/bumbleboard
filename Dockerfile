FROM golang AS builder
WORKDIR /build
COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
    go build -buildvcs=false -a -installsuffix cgo -ldflags="-w -s" -o bbs
FROM scratch AS production
WORKDIR /app
COPY --from=builder /build/static ./static
COPY --from=builder /build/index.html .
COPY --from=builder /build/bbs .
EXPOSE 8080
CMD ["./bbs"]
