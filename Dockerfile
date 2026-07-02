# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install Node.js for frontend build
RUN apk add --no-cache nodejs npm

# Copy backend dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build frontend
RUN cd frontend && npm install && npm run build

# Copy frontend dist to backend
RUN mkdir -p backend/dist && cp -r frontend/dist/* backend/dist/

# Build Go binary
RUN cd backend && go build -o ../server .

# Runtime stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
