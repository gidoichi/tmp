FROM golang:1.22.6-alpine AS builder
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /usr/local/bin/app .

FROM golang:1.22.6-alpine AS admission-webhook
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
WORKDIR /usr/src/app/admission-webhook
RUN go build -v -o /usr/local/bin/app ./cmd

FROM alpine:3.20.1
COPY --from=builder /usr/local/bin/app /usr/local/bin/secrets-store-csi-driver-provider-infisical
COPY --from=admission-webhook /usr/local/bin/app /usr/local/bin/admission-webhook
CMD ["secrets-store-csi-driver-provider-infisical"]
