FROM golang:1.25.5-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/bot ./cmd/bot

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /out/bot /bot
USER nonroot:nonroot
EXPOSE 8090
ENTRYPOINT ["/bot"]