FROM golang:1.27.1-alpine AS build
# Versión exacta = la de go.mod. Evita que Go intente descargar otro toolchain
# en el builder (GOTOOLCHAIN=auto falla si el builder no tiene red hacia el proxy).
ENV GOTOOLCHAIN=local
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY data/ /src/data/
# Pasos separados: si uno falla, el log muestra cuál binario rompió.
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/seed ./cmd/seed

FROM alpine:3.20
# Sin esto, las llamadas HTTPS a Gemini fallan con x509 cuando GEMINI_API_KEY está configurado.
RUN apk add --no-cache ca-certificates
WORKDIR /app/backend
COPY --from=build /out/api /out/seed ./
COPY --from=build /src/backend/migrations ./migrations
COPY --from=build /src/data /app/data
EXPOSE 8080
CMD ["./api"]
