#!/bin/sh

### N3iwf IPSec tunnel configuration
ip a add 192.168.127.1/24  dev eth0
ip link add name ipsec0 type vti local 192.168.127.1 remote 0.0.0.0 key 5
ip a add 10.0.0.1/24 dev ipsec0
ip l set dev ipsec0 up
ip ro del default

