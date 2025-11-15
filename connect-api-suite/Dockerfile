# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json bun.lockb ./

# Install dependencies
RUN npm install -g bun && bun install

# Copy source code
COPY . .

# Build the application
RUN bun run build

# Runtime stage - Nginx to serve the built files
FROM nginx:alpine

# Copy built files from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/ || exit 1

# Run nginx
CMD ["nginx", "-g", "daemon off;"]
