# LTE Controller

A simple controller to manage an LTE WAN connection on a software router.

## Modem manager command reference

```shell
# Extract modem index.
index=$(mmcli --list-modems | grep -oP '\/Modem\/\K\d+')
# Show modem info.
mmcli -m $index
# Enable modem.
mmcli -m $index --simple-connect='apn=bredband.oister.dk,ip-type=ipv4v6'
# Extract bearer index.
bearer=$(mmcli -m $index -K | grep -oP '\/Bearer\/\K\d+')

# Extract the IPv4 configuration.
interface=$(mmcli -m "$index" --bearer "$bearer" -K | grep -oP 'bearer\.status\.interface\s*:\s*\K\S+')
ipv4_addr=$(mmcli -m "$index" --bearer "$bearer" -K | grep -oP 'bearer\.ipv4-config\.address\s*:\s*\K\S+')
ipv4_prefix_len=$(mmcli -m "$index" --bearer "$bearer" -K | grep -oP 'bearer\.ipv4-config\.prefix\s*:\s*\K\S+')
ipv4_mtu=$(mmcli -m "$index" --bearer "$bearer" -K | grep -oP 'bearer\.ipv4-config\.mtu\s*:\s*\K\S+')

# Ensure the wwan0 interface is up.
ip link set $interface up
# Add an IP address to the wwan0 interface.
ip addr add $ip/$prefix_len dev $interface
# Disable ARP on the wwan0 interface.
ip link set dev $interface arp off
# Set the MTU on the wwan0 interface.
ip link set dev $interface mtu $mtu
# Add a default route through the wwan0 interface.
ip route add default dev $interface metric 200

# Check connectivity.
curl -sSL https://api.ipify.org
```
