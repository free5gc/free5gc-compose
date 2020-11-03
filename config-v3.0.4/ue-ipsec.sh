#!/bin/sh

remote_IP=10.100.200.2

### UE IPSec tunnel configuration
ip l add name ipsec0 type vti local ue remote $remote_IP key 5

ip l set dev ipsec0 up
sleep 5
