# Development stage with hot reload
FROM node:20-alpine

WORKDIR /app

# Copy package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy all frontend files
COPY frontend/ ./

# Expose port
EXPOSE 3000

# Start Next.js in development mode
CMD ["npm", "run", "dev"]