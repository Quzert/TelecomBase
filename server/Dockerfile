FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/telecombase-api ./cmd/api

FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=build /out/telecombase-api /app/telecombase-api
COPY --from=build /app/db/queries /app/db/queries

ENV API_PORT=8080
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/telecombase-api"]
