FROM golang:1.26-bookworm

# Install dev tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    make \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Set up non-root user (UID 1000 matches default host user on Linux/Mac)
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000

RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && chown -R $USERNAME:$USERNAME /go

USER $USERNAME
WORKDIR /workspace
