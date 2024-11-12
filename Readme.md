# Tibz-Proxy-Server

This proxy server is meant to run locally

#### Find network interfaces

```.sh

ip a

```

#### Enable IP Forwarding

To route traffic between interfaces, enable IP forwarding:

```.sh

sudo sysctl -w net.ipv4.ip_forward=1

```

#### To make this setting persistent across reboots, add it to /etc/sysctl.conf

```.sh

echo "net.ipv4.ip_forward = 1" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p


```

#### Set Up `iptables` Rules for Transparent Proxying

Use iptables to redirect all HTTP (port 80) and HTTPS (port 443) traffic from devices connected to your hotspot through your proxy server.

```.sh

# Redirect HTTP traffic (port 80) to the proxy server port (e.g., 8080)
sudo iptables -t nat -A PREROUTING -i wlan0 -p tcp --dport 80 -j REDIRECT --to-port 8080

# Redirect HTTPS traffic (port 443) to the proxy server port (e.g., 8080)
sudo iptables -t nat -A PREROUTING -i wlan0 -p tcp --dport 443 -j REDIRECT --to-port 8080

```

#### Optional: Persist iptables Rules

To keep the rules after a reboot, save them:

```.sh

sudo iptables-save | sudo tee /etc/iptables/rules.v4

```
