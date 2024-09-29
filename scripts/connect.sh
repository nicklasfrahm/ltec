#!/usr/bin/env bash
# Usage: ./connect.sh

APN="bredband.oister.dk"

# Extract modem index.
echo ">> Listing modems ..."
index=$(mmcli --list-modems | grep -oP '\/Modem\/\K\d+')

echo ">> Checking modem status ..."
if ! mmcli -m "$index" -K | grep -oE '\/Bearer\/\K\d+'; then
    echo ">> Connecting modem $index to $APN ..."
    if ! mmcli -m "$index" --simple-connect "apn=$APN,ip-type=ipv4v6" >/dev/null; then
        echo "error: failed to connect modem"
        exit 1
    fi

    echo ">> Connected modem $index to $APN"
fi

echo ">> Detecting bearer ..."
bearer=$(mmcli -m "$index" -K | grep -oP '\/Bearer\/\K\d+')

echo ">> Extracting IP configuration ..."
interface=$(mmcli -m "$index" --bearer "$bearer" -K | grep -oP 'bearer\.status\.interface\s*:\s*\K\S+')
ipv4_addr=$(mmcli -m "$index" --bearer "$bearer" -K | grep -oP 'bearer\.ipv4-config\.address\s*:\s*\K\S+')
ipv4_prefix_len=$(mmcli -m "$index" --bearer "$bearer" -K | grep -oP 'bearer\.ipv4-config\.prefix\s*:\s*\K\S+')
ipv4_mtu=$(mmcli -m "$index" --bearer "$bearer" -K | grep -oP 'bearer\.ipv4-config\.mtu\s*:\s*\K\S+')

echo ">> Checking if interface $interface is up ..."
if ! ip addr show dev "$interface" | grep -q "$ipv4_addr"; then
    echo ">> Configuring network interface ..."
    ip link set "$interface" up
    ip addr add "$ipv4_addr/$ipv4_prefix_len" dev "$interface"
    ip link set dev "$interface" arp off
    ip link set dev "$interface" mtu "$ipv4_mtu"
    ip route add default dev "$interface" metric 200
fi

echo ">> Checking connectivity ..."
if ! curl -sSL api.ipify.org >/dev/null; then
    echo "error: unable to obtain public IP"
fi

public_ip=$(curl -sSL api.ipify.org)
echo ">> Successfully connected using public IP: $public_ip"
