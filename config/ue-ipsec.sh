#!/bin/sh

### UE IPSec tunnel configuration
ip route del default # Docker adds a default route to the gateway and it conflicts with the route UE adds to GRE tunnel
ip l add name ipsec0 type vti local 192.168.127.2 remote 192.168.127.1 key 5
ip l set dev ipsec0 up
./ue -uecfg ../config/uecfg.conf