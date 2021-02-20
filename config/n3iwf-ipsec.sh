#!/bin/sh

### N3iwf IPSec tunnel configuration

# As per https://github.com/free5gc/free5gc/issues/45#issuecomment-634012712
#     IKEBindAddress: dynamically computed by $(hostname -i | awk '{print $1}')
#     IPSecInterfaceMark: 5
#     IPSecInterfaceAddress: 10.0.0.1
#     IPSec subnet CIDR: /24
#     N3IWF tunnel interface: ipsec0
#     

ip link add name ipsec0 type vti local $(hostname -i | awk '{print $1}') remote 0.0.0.0 key 5
ip addr add 10.0.0.1/24 dev ipsec0
ip link set dev ipsec0 up
