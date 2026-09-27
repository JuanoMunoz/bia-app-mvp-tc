FROM golang:1.27-alpine AS build
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY data/ /src/data/
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api && CGO_ENABLED=0 go build -o /out/seed ./cmd/seed

FROM alpine:3.20
WORKDIR /app/backend
COPY --from=build /out/api /out/seed ./
COPY --from=build /src/backend/migrations ./migrations
COPY --from=build /src/data /app/data
EXPOSE 8080
CMD ["./api"]
