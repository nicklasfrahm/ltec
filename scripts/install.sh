#!/usr/bin/env bash
# Usage: curl -sSL https://raw.githubusercontent.com/nicklasfrahm/ltec/main/scripts/install.sh | sudo bash -s -- <apn>

set -eou pipefail

REPO="nicklasfrahm/ltec"
GREEN='\033[0;32m'
RED='\033[0;31m'
RESET='\033[0m'

print_error() {
  echo -e "${RED}err:${RESET} $1"
}

print_info() {
  echo -e "${GREEN}inf:${RESET} $1"
}

main() {
  APN="$1"
  if [ -z "$APN" ]; then
    print_error "missing required argument: <access_point_name>"
    exit 1
  fi

  # Check if script is running as root.
  if [ "$EUID" -ne 0 ]; then
    print_error "script must be run as root or using sudo"
    exit 1
  fi

  if dpkg -l modemmanager >/dev/null 2>&1; then
    print_info "Installing ModemManager ..."
    apt-get update && apt-get install -y modemmanager
  fi

  arch="amd64"
  if [ "$(uname -m)" == "aarch64" ]; then
    arch="arm64"
  fi

  print_info "Stopping existing ltec service ..."
  systemctl stop ltec.service || true

  print_info "Downloading and installing ltec binary ..."
  curl -sSL "https://github.com/$REPO/releases/latest/download/ltec-linux-${arch}" -o /usr/bin/ltec
  chmod +x /usr/bin/ltec

  # Create service file.
  print_info "Creating systemd service ..."
  export APN="$APN"
  curl -sSL https://raw.githubusercontent.com/$REPO/main/configs/systemd/ltec.service | envsubst >/etc/systemd/system/ltec.service

  print_info "Reloading systemd daemons ..."
  systemctl daemon-reload

  print_info "Enabling and starting ltec service ..."
  systemctl enable --now ltec.service
  systemctl restart ltec.service
}

# We want to pass all arguments to main.
#shellcheck disable=SC2068
main $@
