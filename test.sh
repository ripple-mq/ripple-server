#!/bin/bash

echo "Starting Zookeeper..."
docker-compose -f test.compose.yml up  -d zookeeper

echo "Waiting for Zookeeper to initialize..."
sleep 10 

echo "Building and starting ripple-server..."
docker-compose -f test.compose.yml up  --build ripple-server-1 ripple-server-2 ripple-server-3

echo "Checking the status of the containers..."
docker-compose ps

echo "Script execution completed."
