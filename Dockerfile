FROM golang:1.22.6-alpine AS builder
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /usr/local/bin/app ./main.go

FROM alpine:3.20.1
COPY --from=builder /usr/local/bin/app /usr/local/bin/secrets-store-csi-driver-provider-infisical
CMD ["secrets-store-csi-driver-provider-infisical"]
