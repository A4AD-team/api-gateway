#!/bin/bash
# docker/rabbitmq/generate-config.sh

set -e

echo "Generating RabbitMQ configuration from environment variables..."

# Create definitions.json from environment variables
cat > /etc/rabbitmq/definitions.json <<EOF
{
  "users": ${RABBITMQ_USERS:-[]},
  "vhosts": [{"name": "/"}],
  "permissions": ${RABBITMQ_PERMISSIONS:-[]},
  "queues": ${RABBITMQ_QUEUES:-[]},
  "exchanges": ${RABBITMQ_EXCHANGES:-[]},
  "bindings": ${RABBITMQ_BINDINGS:-[]},
  "policies": ${RABBITMQ_POLICIES:-[]}
}
EOF

echo "Configuration generated successfully!"