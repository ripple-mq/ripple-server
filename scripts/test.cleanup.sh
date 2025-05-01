#!/bin/bash

echo "Stopping and removing ripple-server and zookeeper containers..."
docker-compose -f test.compose.yml down --remove-orphans

echo "Removing custom Docker network..."
docker network rm custom-net 2>/dev/null || echo "Custom network not found."



echo "Cleanup completed."
