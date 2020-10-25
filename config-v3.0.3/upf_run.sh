#!/bin/bash

sleep 3 

# Enable masquerade routing
iptables -t nat -A POSTROUTING -o ${DN_NETWORK} -j MASQUERADE

# Add test network
ip addr add ${IP_TEST} dev lo

# up dev lo
ip link set up dev lo 

#set up tunnel mtu
ip link set dev upfgtp0 mtu 1500
