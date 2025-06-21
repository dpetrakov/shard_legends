# Internal API Gateway for microservices
FROM nginx:alpine

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Create log directory
RUN mkdir -p /var/log/nginx

# Set proper permissions
RUN chown -R nginx:nginx /var/log/nginx

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]