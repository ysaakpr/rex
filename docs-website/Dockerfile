# Multi-stage Dockerfile for VitePress Documentation

# Development stage - with hot reload
FROM node:20-alpine AS development
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy the rest of the application
COPY . .

# Expose VitePress dev server port
EXPOSE 5173

# Start VitePress dev server
CMD ["npm", "run", "docs:dev", "--", "--host", "0.0.0.0", "--port", "5173"]

# Build stage - builds static site
FROM node:20-alpine AS builder
WORKDIR /app

# Build argument for demo URL
ARG VITE_DEMO_URL=http://localhost/demo
ENV VITE_DEMO_URL=$VITE_DEMO_URL

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source files
COPY . .

# Build the documentation site
RUN npm run docs:build

# Production stage - serve static files with nginx
FROM nginx:alpine AS production
WORKDIR /usr/share/nginx/html

# Remove default nginx static assets
RUN rm -rf ./*

# Copy built static files from builder stage
COPY --from=builder /app/docs/.vitepress/dist .

# Copy custom nginx configuration if needed
# COPY nginx.conf /etc/nginx/nginx.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]

