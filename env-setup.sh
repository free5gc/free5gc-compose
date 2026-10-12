#!/bin/bash

GO_VERSION="1.26.2"
GOLANGCI_LINT_VERSION="2.11.4"

NETWORK_INTERFACE=""
IP_ADDRESS=""
LINT=false
DOCKER=false

GTP5G_PATH="$HOME/gtp5g"
GTP5G_VERSION="v0.9.16"

SUCCESS_COUNT=0
FAIL_COUNT=0
SKIP_COUNT=0

COLOR_RED="\033[31m"
COLOR_GREEN="\033[32m"
COLOR_YELLOW="\033[33m"
COLOR_BLUE="\033[36m"
COLOR_RESET="\033[0m"

log_info() {
    echo -e "${COLOR_BLUE}[.]${COLOR_RESET} $1"
}

log_success() {
    echo -e "${COLOR_GREEN}[+]${COLOR_RESET} $1"
}

log_warn() {
    echo -e "${COLOR_YELLOW}[!]${COLOR_RESET} $1"
}

log_question() {
    echo -e "${COLOR_YELLOW}[?]${COLOR_RESET} $1"
}

log_error() {
    echo -e "${COLOR_RED}[-]${COLOR_RESET} $1"
}

separate_stars() {
    local cols=$(tput cols 2>/dev/null || echo 10)
    printf "%*s\n" "$cols" "" | tr ' ' '*'
}

usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help                            Show this help message"
    echo
    echo "Example:"
    echo "  $0"
    echo
}

install_gtp5g() {
    log_info "Installing gtp5g..."

    if lsmod | grep -q gtp5g; then
        log_info "GTP5G already installed"
        SKIP_COUNT=$((SKIP_COUNT + 1))
        return
    fi

    sudo apt -y update
    sudo apt -y install gcc g++ cmake autoconf libtool pkg-config libmnl-dev libyaml-dev

    git clone --branch ${GTP5G_VERSION} https://github.com/free5gc/gtp5g.git $GTP5G_PATH
    pushd $GTP5G_PATH
    make
    sudo make install
    popd

    log_success "GTP5G installed"
    SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
}

install_docker() {
    log_info "Installing Docker..."

    if docker --version > /dev/null 2>&1; then
        log_info "Docker already installed"
        SKIP_COUNT=$((SKIP_COUNT + 1))
        return
    fi

    sudo apt update
    sudo apt install ca-certificates curl
    sudo install -m 0755 -d /etc/apt/keyrings
    sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    sudo chmod a+r /etc/apt/keyrings/docker.asc

    sudo tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Signed-By: /etc/apt/keyrings/docker.asc
EOF

    sudo apt update
    sudo apt install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    sudo groupadd docker
    sudo usermod -aG docker $USER

    log_success "Docker installed"
    SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
}

print_counts() {
    log_info "Task Summary:"
    log_success "  Success: $SUCCESS_COUNT"
    log_warn "  Skip: $SKIP_COUNT"
    log_error "  Fail: $FAIL_COUNT"
}

main() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                return 0
                ;;
            *)
                usage
                return 1
                ;;
        esac
    done

    install_gtp5g
    separate_stars
    
    install_docker
    separate_stars

    print_counts
    separate_stars

    if $DOCKER; then
        log_info "Docker is installed. Please log out and log in again to use Docker without sudo."
    fi
}

main "$@"