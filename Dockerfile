FROM golang:1.26.11 AS build
WORKDIR /src
COPY go.mod ./
COPY . .
ARG BRAIN_COVERAGE=0
RUN if [ "$BRAIN_COVERAGE" = 1 ]; then \
      CGO_ENABLED=0 go build -trimpath -cover -covermode=atomic -coverpkg=./internal/...,./cmd/... -o /out/brain ./cmd/brain; \
    else \
      CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/brain ./cmd/brain; \
    fi

FROM node:20-bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends bubblewrap bash && rm -rf /var/lib/apt/lists/*
WORKDIR /workspace
COPY package.json package-lock.json ./
RUN npm ci --ignore-scripts
COPY tsconfig.json ./
COPY scripts/typescript-worker.mjs scripts/function-worker.mjs scripts/run-sandboxed-typescript-worker.sh ./scripts/
COPY plugins ./plugins
COPY --from=build /out/brain /brain
RUN mkdir -p /workspace/.local/coverage && chown -R node:node /workspace/.local
USER node
EXPOSE 8080
ENTRYPOINT ["/brain"]
