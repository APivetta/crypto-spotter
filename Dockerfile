FROM golang:1.23.4

WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY src/ ./src

# Install dependencies
RUN apt update && apt install -y cron

# Add cron job
RUN echo "0 23 * * * /app/bin/fetch && /app/bin/train && /app/bin/trader >> /var/log/cron.log 2>&1" | crontab -

# Ensure cron runs in the background when the container starts
RUN mkdir -p /var/spool/cron/crontabs && chmod 600 /var/spool/cron/crontabs

# Start cron in the background
RUN echo '#!/bin/bash\ncron &\nexec "$@"' > /start.sh && chmod +x /start.sh

# Build the application
RUN go build -o bin/trader ./src
RUN go build -o bin/train ./src/cmd/train/
RUN go build -o bin/fetch ./src/cmd/fetch_snapshots/

# Start cron in the background
RUN echo '#!/bin/bash\ncron &\n sh -c "$@"' > /start.sh && chmod +x /start.sh

# Set the command to run cron and the application
CMD ["/start.sh", "/app/bin/fetch && /app/bin/train && /app/bin/trader"]
