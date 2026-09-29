# syntax=docker/dockerfile:1
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
COPY web ./web
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/mm-inches .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/mm-inches /mm-inches
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/mm-inches"]
