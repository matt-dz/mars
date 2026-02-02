FROM --platform=$BUILDPLATFORM golang:1.25.1-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o ./bin/mars ./cmd/mars/main.go

FROM alpine:latest AS final
WORKDIR /app
COPY --from=build /app/bin/mars /app/mars
EXPOSE 8080
ENTRYPOINT ["/app/mars"]
