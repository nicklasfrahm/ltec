#!/usr/bin/env bash
# Usage: curl -sSL https://raw.githubusercontent.com/nicklasfrahm/ltec/main/scripts/install.sh | sudo bash -s -- <apn>

set -eou pipefail

repo="nicklasfrahm/ltec"

main() {
  APN="$1"
  if [ -z "$APN" ]; then
    echo "error: missing required argument: <access_point_name>"
    exit 1
  fi

  # Check if script is running as root.
  if [ "$EUID" -ne 0 ]; then
    echo "error: script must be run as root or using sudo"
    exit 1
  fi

  # Install dependencies.
  apt-get update
  apt-get install -y modemmanager

  # Install binary.
  curl -sSL "https://github.com/$repo/releases/download/latest/ltec" -o /usr/bin/ltec

  # Create service file.
  export APN="$APN"
  curl -sSL https://raw.githubusercontent.com/$repo/main/configs/systemd/ltec.service | envsubst >/etc/systemd/system/ltec.service

  # Reload systemd.
  systemctl daemon-reload

  # Enable and start service.
  systemctl enable --now ltec.service
}

# We want to pass all arguments to main.
#shellcheck disable=SC2068
main $@
