# Build stage: compile a static binary (pgx driver is pure Go, so CGO is off).
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/tasker-web ./cmd/tasker-web

# Runtime stage: distroless static — no shell, no package manager, minimal
# attack surface. Holds nothing but the binary.
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/tasker-web /tasker-web
EXPOSE 5001
ENTRYPOINT ["/tasker-web"]
