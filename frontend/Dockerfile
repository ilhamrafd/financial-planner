# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm ci

# Build Vue app
COPY . .
RUN npm run build

# Runtime stage – Nginx
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html

# Optional: custom Nginx config (use default if not provided)
#COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]