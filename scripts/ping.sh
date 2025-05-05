#!/bin/bash

# Check for correct arguments
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <host> <port>"
    exit 1
fi

HOST=$1
PORT=$2

# Infinite loop to keep pinging the port
while true; do
    nc -zv -w 1 "$HOST" "$PORT" > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo "$(date): Port $PORT on $HOST is OPEN."
    else
        echo "$(date): Port $PORT on $HOST is CLOSED."
    fi
    sleep 1
done
