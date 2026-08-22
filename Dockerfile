FROM node:24-alpine AS assets

WORKDIR /app/web

COPY web/package.json web/package-lock.json ./
RUN npm install

COPY web/ ./
RUN npm run build


FROM golang:1.27.0-alpine AS builder

WORKDIR /app

ARG VERSION=dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=assets /app/internal/web/dist ./internal/web/dist

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=${VERSION}" -o /tmp/watchue ./cmd/web


FROM gcr.io/distroless/static-debian13:latest

COPY --from=builder /tmp/watchue /watchue

EXPOSE 8080

CMD ["/watchue"]
