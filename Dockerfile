FROM node:22-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.25-alpine AS backend
WORKDIR /src/backend
RUN apk add --no-cache ca-certificates
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o /out/biu-panel ./cmd/server

FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend /out/biu-panel /app/biu-panel
COPY --from=frontend /src/frontend/dist /app/public
RUN mkdir -p /app/data
ENV BIU_PANEL_PORT=55088
ENV BIU_PANEL_DATA_DIR=/app/data
ENV BIU_PANEL_STATIC_DIR=/app/public
EXPOSE 55088
VOLUME ["/app/data"]
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget -qO- http://127.0.0.1:55088/api/health || exit 1
CMD ["/app/biu-panel"]
