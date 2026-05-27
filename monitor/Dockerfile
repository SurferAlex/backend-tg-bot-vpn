FROM golang:1.25.5-alpine AS build
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/monitor ./cmd/monitor

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /out/monitor /monitor
USER nonroot:nonroot
ENTRYPOINT ["/monitor"]
