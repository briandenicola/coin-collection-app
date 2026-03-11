FROM mcr.microsoft.com/devcontainers/go:1.26

ARG NODE_VERSION="24"
RUN su vscode -c "curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | sudo -E bash - && sudo apt-get install -y nodejs"
