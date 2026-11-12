FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/brain ./cmd/brain

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/brain /brain
EXPOSE 8080
ENTRYPOINT ["/brain"]
