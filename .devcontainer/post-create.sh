#!/bin/bash
echo "$(date) post-create start" >> ~/status

# Install Task runner
sudo sh -c "$(curl -sL https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

# Install Go dependencies
cd /workspaces/AncientCoins/src/api
go mod download

# Install Node dependencies
cd /workspaces/AncientCoins/src/web
npm install

echo "$(date) post-create complete" >> ~/status
