#!/usr/bin/bash

src=$(curl -s https://api-ipv6.ip.sb/ip -A Mozilla)
m=$(ip link show $1 | grep -o -E "(([a-fA-F0-9]{2}:){5}[a-fA-F0-9]{2})" | head -n 1)
g=$(ip -6 neigh show | grep router | grep -o -E "(([a-fA-F0-9]{2}:){5}[a-fA-F0-9]{2})" | head -n 1)
echo "Interface : $1, Src Address : $src, Mac : $m , Gateway : $g"
go clean
go build -buildvcs=false -ldflags "-X 'main.ifacename=$1' -X 'main.src=$src' -X 'main.smac=$m' -X 'main.dmac=$g'"
