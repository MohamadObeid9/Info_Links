# --- frontend build ---
FROM node:24-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# --- go build ---
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

# --- runtime ---
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /server /server
COPY --from=builder /app/frontend/dist /frontend/dist
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
