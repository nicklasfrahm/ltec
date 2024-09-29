#!/usr/bin/env bash
# Usage: ./connect.sh

APN="bredband.oister.dk"
GREEN='\033[0;32m'
RED='\033[0;31m'
RESET='\033[0m'

main() {
  echo -e "${GREEN}inf:${RESET} Detecting modems ..."
  index=$(mmcli --list-modems | grep -oP '\/Modem\/\K\d+')
  if [ -z "$index" ]; then
    echo -e "${RED}err:${RESET} No modems found"
    exit 1
  fi

  echo -e "${GREEN}inf:${RESET} Found modem $index"

  bearer=$(mmcli -m "$index" -K | grep -oP '\/Bearer\/\K\d+')
  if [ -z "$bearer" ]; then
    echo -e "${GREEN}inf:${RESET} Connecting modem $index to $APN ..."
    if ! mmcli -m "$index" --simple-connect "apn=$APN,ip-type=ipv4v6" >/dev/null; then
      echo -e "${RED}err:${RESET} Failed to connect modem $index to $APN"
      exit 1
    fi

    echo -e "${GREEN}inf:${RESET} Successfully connected modem $index to $APN"
  fi

  echo -e "${GREEN}inf:${RESET} Detecting bearer ..."
  bearer=$(mmcli -m "$index" -K | grep -oP '\/Bearer\/\K\d+')
  if [ -z "$bearer" ]; then
    echo -e "${RED}err:${RESET} No bearer found"
    exit 1
  fi

  echo -e "${GREEN}inf:${RESET} Found bearer $bearer"

  echo -e "${GREEN}inf:${RESET} Detecting interface ..."
  interface=$(mmcli -m "$index" --bearer "$bearer" -K | grep -oP 'bearer\.status\.interface\s*:\s*\K\S+')
  if [ -z "$interface" ]; then
    echo -e "${RED}err:${RESET} No interface found"
    exit 1
  fi

  echo -e "${GREEN}inf:${RESET} Checking interface configuration ..."
  if ! ip addr show dev "$interface" | grep -q "$ipv4_addr"; then
    echo -e "${GREEN}inf:${RESET} Configuring interface $interface ..."

    ipv4_addr=$(mmcli -m "$index" --bearer "$bearer" -K | grep -oP 'bearer\.ipv4-config\.address\s*:\s*\K\S+')
    ipv4_prefix_len=$(mmcli -m "$index" --bearer "$bearer" -K | grep -oP 'bearer\.ipv4-config\.prefix\s*:\s*\K\S+')
    ipv4_mtu=$(mmcli -m "$index" --bearer "$bearer" -K | grep -oP 'bearer\.ipv4-config\.mtu\s*:\s*\K\S+')

    ip link set "$interface" up
    ip addr add "$ipv4_addr/$ipv4_prefix_len" dev "$interface"
    ip link set dev "$interface" arp off
    ip link set dev "$interface" mtu "$ipv4_mtu"
    ip route add default dev "$interface" metric 200
  fi

  echo -e "${GREEN}inf:${RESET} Successfully configured interface $interface"

  echo -e "${GREEN}inf:${RESET} Detecting public IP ..."
  if ! curl -sSL api.ipify.org >/dev/null; then
    echo -e "${RED}err:${RESET} Failed verify internet connection"
    exit 1
  fi

  public_ip=$(curl -sSL api.ipify.org)
  echo -e "${GREEN}inf:${RESET} Successfully connected to internet with public IP $public_ip"
}

# We want to pass all arguments to main.
#shellcheck disable=SC2068
main $@
