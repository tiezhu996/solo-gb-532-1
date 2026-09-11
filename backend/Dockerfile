FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget tzdata
WORKDIR /app
COPY --from=build /out/server /app/server
EXPOSE 8080
USER nobody
ENTRYPOINT ["/app/server"]
