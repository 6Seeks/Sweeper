package main

import (
  "encoding/binary"
  "flag"
  "fmt"
  "math"
  "math/rand"

  "crypto/cipher"
  "crypto/des"
  "log"
  "net"
  _ "net/http/pprof"
  "os"
  "runtime"
  "sync/atomic"
  "time"
  "unsafe"
)

const instance uint16 = 0x1314
const key = "12345678"

const tau = 0.1

type PCS struct {
  prefix     uint64
  acceptance float64
  numProbe   uint32
  numRouter  uint32
  rejectIdx  int
}

var (
  ifacename string
  src       string
  smac      string
  dmac      string
  minTTL    int
  maxTTL    int
  count     int
  pain      int     = 0
  gain      int     = 0
  sumWeight float64 = 0

  s float64

  PCSList            = make([]PCS, 1<<24) // not dynamci allocation
  PCScount           = 0
  BitSet             = make([]byte, 1<<29) // bloom filter
  sender   *Transmit = new(Transmit)
  block    cipher.Block
)

func murmur3(data []byte, seed uint32) uint32 {
  hash := seed

  for i := 0; i < len(data); i = i + 4 {
    k := binary.BigEndian.Uint32(data[i : i+4])
    k = k * 0xcc9e2d51
    k = (k << 15) | (k >> 17)
    k = k * 0x1b873593
    hash = hash ^ k
    hash = (hash << 13) | (hash >> 19)
    hash = hash*5 + 0xe6546b64
  }
  hash = hash ^ (hash >> 16)
  hash = hash * 0x85ebca6b
  hash = hash ^ (hash >> 13)
  hash = hash * 0xc2b2ae35
  hash = hash ^ (hash >> 16)
  return hash
}

func fnv1a(value uint32) uint32 {
  var hash uint32 = 2166136261
  for i := 0; i < 4; i++ {
    hash ^= value & 0xff
    hash *= 16777619
    value >>= 8
  }
  return hash
}

func LoadUint16(addr *uint16) uint16 {
  return uint16(uintptr(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(addr)))))
}

func main() {
  flag.IntVar(&minTTL, "l", 4, "")
  flag.IntVar(&maxTTL, "m", 32, "")
  flag.IntVar(&count, "c", 1e9, "")
  flag.Parse()
  rand.Seed(time.Now().UnixNano())

  block, _ = des.NewCipher([]byte(key))
  logfile, err := os.OpenFile("logfile", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)

  if err != nil {
    log.Fatal(err)
  }

  defer logfile.Close() // 关闭文件
  log.SetOutput(logfile)

  for {
    var line string
    if _, err := fmt.Scanln(&line); err != nil {
      break
    }
    if _, ip6net, err := net.ParseCIDR(line); err != nil {
      panic(line)
    } else {
      l, _ := ip6net.Mask.Size()
      if l != 48 {
        panic("Illegal Prefix:" + line)
      }
      PCSList[PCScount].prefix = binary.BigEndian.Uint64(ip6net.IP[:8])
      PCScount++
    }
  }
  PCSList = PCSList[:PCScount]
  runtime.GC() //
  fmt.Println(PCScount, "/48s")
  log.Printf("%d /48s\n", PCScount)
  sender.Init()
  go Capture()
  fmt.Println("Starting...")
  var reward, offset float64
  for count > 0 {
    for i := 0; i < PCScount; i++ {
      reward = float64(atomic.LoadUint32(&PCSList[i].numRouter))
      offset = float64(PCSList[i].numProbe) + 1e-6
      PCSList[i].acceptance = math.Exp2(reward / offset / tau) 
      if offset <= reward {
        PCSList[i].acceptance = 0.0
      }
      sumWeight += PCSList[i].acceptance
    }

    // scale acceptance
    for i := 0; i < PCScount; i++ {
      PCSList[i].acceptance /= sumWeight
      PCSList[i].acceptance *= float64(PCScount)
    }
    FlushAreaDivision()
    s = rand.Float64()
    if count > 1e7 {
      Probe(1e7)
      count -= 1e7
    } else {
      Probe(count)
      count = 0
    }
    time.Sleep(10 * time.Second)
    fmt.Printf("%s %d/%d\n", time.Now().Format("20060102-150405"), gain, pain)
    log.Printf("%s %d/%d\n", time.Now().Format("20060102-150405"), gain, pain)
  }

}
