#!/bin/sh

### N3iwf IPSec tunnel configuration
ip link add name ipsec0 type vti local 127.0.0.1 remote 0.0.0.0 key 5
ip a add 10.0.0.1/24 dev ipsec0
ip l set dev ipsec0 up
ip ro del default

