import ipaddress
import sys

with open(sys.argv[1]) as f:
    for line in f.read().splitlines():
        n = ipaddress.IPv6Network(line)
        for line in n.subnets(new_prefix=48):
            print(line)


