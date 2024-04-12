package main

import (
  "math/rand"
  "runtime"
  "math"
)

func FlushAreaDivision() {
  // two pipes
  small := make([]int, PCScount)
  large := make([]int, PCScount)
  smallCount, largeCount := 0, 0
  for i := 0; i < PCScount; i++ {
    if PCSList[i].acceptance < 1.0 { // for lower memory  usage
      small[smallCount] = i
      smallCount++
    } else {
      large[largeCount] = i
      largeCount++
    }
  }
  small = small[:smallCount]
  large = large[:largeCount]
  runtime.GC() // force garbarge collection the slices small and large
  var s, l int
  for len(small) > 0 && len(large) > 0 {
    s = small[0]
    small = small[1:]
    l = large[0]
    large = large[1:]
    PCSList[s].rejectIdx = l
    PCSList[l].acceptance = PCSList[l].acceptance + PCSList[s].acceptance - 1.0
    if PCSList[l].acceptance < 1.0 {
      small = append(small, l)
    } else {
      large = append(large, l)
    }
  }
  for len(small) > 0 {
    s = small[0]
    small = small[1:]
    PCSList[s].rejectIdx = s
  }
  for len(large) > 0 {
    l = large[0]
    large = large[1:]
    PCSList[l].rejectIdx = l
  }
  runtime.GC() // force garbarge collection the slices small and large
}

func Generate() int {
  column := rand.Intn(PCScount)
  if PCSList[column].acceptance < rand.Float64() {
    return PCSList[column].rejectIdx
  }
  return column
}

func evenGenerate() int {
  s = math.Mod(s+0.6180339887498949, 1.0)
  column := math.Floor(float64(PCScount) * s)
  if PCSList[int(column)].acceptance < float64(PCScount)*s-column {
    return PCSList[int(column)].rejectIdx
  }
  return int(column)
}
// func main() {
//  myat := &AliasTable{
//    prob:   []float32{0.1 * 5, 0.4 * 5, 0.15 * 5, 0.3 * 5, 0.05 * 5},
//    accept: []uint32{7, 8, 9, 12, 14},
//    reject: []uint32{0, 0, 0, 0, 0},
//    size:   5,
//  }
//  myat.FlushAliasTable()
//
//  for i := 0; i < 10000000; i++ {
//    fmt.Println(myat.Generate())
//  }
//
// }
