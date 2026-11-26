FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/brain ./cmd/brain

FROM node:20-bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends bubblewrap bash && rm -rf /var/lib/apt/lists/*
WORKDIR /workspace
COPY package.json package-lock.json ./
RUN npm ci --ignore-scripts
COPY tsconfig.json ./
COPY scripts/typescript-worker.mjs scripts/run-sandboxed-typescript-worker.sh ./scripts/
COPY plugins ./plugins
COPY --from=build /out/brain /brain
RUN mkdir -p /workspace/.local && chown node:node /workspace/.local
USER node
EXPOSE 8080
ENTRYPOINT ["/brain"]
