package main

import (
  "encoding/binary"
  "fmt"
  "github.com/google/gopacket"
  "github.com/google/gopacket/pcap"
  "golang.org/x/net/ipv6"
  "os"
  "time"
)

func Capture() {
  var decrypted = make([]byte, 8)
  var ins, sts, rts uint16
  var state, hopLimit uint32
  var i int
  // 打开 pcap 文件
  fmt.Println(ifacename)
  handle, err := pcap.OpenLive(ifacename, 1600, false, pcap.BlockForever)
  if err != nil {
    panic(err)
  }
  defer handle.Close()

  // 设置过滤器（可选）
  err = handle.SetBPFFilter("ip6 and icmp6 and icmp6[0]<128")
  if err != nil {
    panic(err)
  }
  file, err := os.Create("output/" + time.Now().Format("20060102-150405"))
  if err != nil {
    panic(err)
  }
  defer file.Close()
  // 从 pcap 文件中循环读取数据包
  packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
  for packet := range packetSource.Packets() {
    rts = uint16(time.Now().UnixMilli())
    ip6header, err := ipv6.ParseHeader(packet.Data()[14:])
    if err != nil {
      panic(err)
    }
    icmp6type := packet.Data()[54]
    icmp6code := packet.Data()[55]
    ip6headerp, err := ipv6.ParseHeader(packet.Data()[62:])
    if err != nil {
      continue
    }
    target := ip6headerp.Dst
    hop := ip6header.Src

    a := murmur3(hop, 0x12345678)
    b := murmur3(hop, 0x87654321)

    if BitSet[a/8]&(1<<(a%8)) != 0 && BitSet[b/8]&(1<<(b%8)) != 0 {
      continue
    }
    BitSet[a/8] |= (1 << (a % 8))
    BitSet[b/8] |= (1 << (b % 8))

    block.Decrypt(decrypted, target[8:16])
    state = binary.BigEndian.Uint32(decrypted[0:4])
    ins = uint16(state >> 16)
    sts = uint16(state)
    state = binary.BigEndian.Uint32(decrypted[4:8])
    i = int(state & 0xff_ffff)
    hopLimit = state >> 24

    fmt.Fprintf(file, "%s ", target) // target
    fmt.Fprintf(file, "%s ", hop)    // hop
    fmt.Fprintf(file, "%d ", icmp6type)
    fmt.Fprintf(file, "%d ", icmp6code)
    fmt.Fprintf(file, "%d ", hopLimit) //
    fmt.Fprintf(file, "%d ", ip6headerp.HopLimit)
    fmt.Fprintf(file, "%d ", ip6header.HopLimit)
    fmt.Fprintf(file, "%d ", rts-sts)
    fmt.Fprintln(file, ins == instance) // the target IP has not been modified

    if murmur3(hop, 0x114514) == murmur3(target, 0x114514) { // ICMP rate limiting
      continue
    }
    if ins == instance {
      PCSList[i].numRouter++
      gain++
    }
  }
}
