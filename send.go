package main

import (
  "encoding/binary"
  "math/rand"
  "net"
  "time"

  "github.com/google/gopacket"
  "github.com/google/gopacket/layers"
  "golang.org/x/sys/unix"
)

type Transmit struct {
  eth       layers.Ethernet
  ip6       layers.IPv6
  icmp6     layers.ICMPv6
  icmp6echo layers.ICMPv6Echo
  payload   gopacket.Payload
  opts      gopacket.SerializeOptions
  fd        int
}

func (tm *Transmit) Init() {
  SrcMAC, err := net.ParseMAC(smac)
  if err != nil {
    panic(err)
  }
  DstMAC, err := net.ParseMAC(dmac)
  if err != nil {
    panic(err)
  }
  SrcIP := net.ParseIP(src)
  tm.eth = layers.Ethernet{EthernetType: layers.EthernetTypeIPv6, SrcMAC: SrcMAC, DstMAC: DstMAC}
  tm.ip6 = layers.IPv6{Version: 6, NextHeader: layers.IPProtocolICMPv6, SrcIP: SrcIP, DstIP: net.IPv6zero}
  tm.icmp6 = layers.ICMPv6{TypeCode: layers.CreateICMPv6TypeCode(128, 0)}

  tm.opts = gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
  tm.icmp6echo = layers.ICMPv6Echo{Identifier: uint16(rand.Int()), SeqNumber: uint16(rand.Int())}
  tm.payload = gopacket.Payload([]byte("https://6Seeks.github.io/"))
  tm.fd, err = unix.Socket(unix.AF_PACKET, unix.SOCK_RAW, ((unix.ETH_P_ALL<<8)&0xff00)|unix.ETH_P_ALL>>8)
  if err != nil {
    panic(err)
  }
  iface, err := net.InterfaceByName(ifacename)
  if err != nil {
    panic(err)
  }
  if err = unix.Bind(tm.fd, &unix.SockaddrLinklayer{Ifindex: iface.Index}); err != nil {
    panic(err)
  }
}
func Probe(steps int) {
  var (
    plaintext = make([]byte, 8)
    buffer    = gopacket.NewSerializeBuffer()
  )
  var i int
  var state uint32 //
  var target uint64
  for steps > 0 {
    i = evenGenerate()
    PCSList[i].numProbe++
    steps--

    // /64
    state = fnv1a(PCSList[i].numProbe + uint32(i))
    sender.ip6.HopLimit = uint8(state%uint32(maxTTL-minTTL+1)) + uint8(minTTL) // HopLimit
    target = PCSList[i].prefix
    target += uint64(state/uint32(maxTTL-minTTL+1)) & 0xffff // hash 48-64 bit, fuck, i am a idoit
    binary.BigEndian.PutUint64(sender.ip6.DstIP[:8], target)

    // IID
    state = uint32(instance) << 16                   // 0 - 16 bits
    state += uint32(time.Now().UnixMilli() & 0xffff) // 16 - 32 bits
    binary.BigEndian.PutUint32(plaintext[0:4], state)

    state = uint32(sender.ip6.HopLimit) << 24 // 32 - 40 bits
    state += uint32(i & 0xff_ffff)            // 40 - 64 bits
    binary.BigEndian.PutUint32(plaintext[4:8], state)

    // fmt.Printf("%x %x %x %d\n", PCSList[i].prefix, target, plaintext, sender.ip6.HopLimit)
    block.Encrypt(sender.ip6.DstIP[8:16], plaintext)

    sender.icmp6echo.Identifier++
    sender.icmp6echo.SeqNumber++
    sender.icmp6.SetNetworkLayerForChecksum(&sender.ip6)
    gopacket.SerializeLayers(buffer,
      sender.opts,
      &sender.eth,
      &sender.ip6,
      &sender.icmp6,
      &sender.icmp6echo,
      &sender.payload)
    if err := unix.Send(sender.fd, buffer.Bytes(), unix.MSG_WAITALL); err != nil {
      panic(err)
    }
    pain++
  }
}
