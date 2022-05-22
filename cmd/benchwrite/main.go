package main

import (
	"fmt"
	"github.com/gballet/go-verkle"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	f, _ := os.Create("cpu.prof")
	g, _ := os.Create("mem.prof")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	defer pprof.WriteHeapProfile(g)

	desc := `
This is a benchmarker which writes key/values to a verkle tree. 
- The keys are 32 bytes, 
- The values are 32 bytes, 

It outputs the root committment after writing 5300 kv-pairs. 
It aborts after 10 seconds, writing cpu.prof and mem.prof files to disk.
`
	fmt.Println(desc)

	tree := verkle.New()
	start := time.Now()
	k := make([]byte, 32)
	v := make([]byte, 32)

	cp := func(src []byte) []byte {
		x := make([]byte, len(src))
		copy(x, src)
		return x
	}
	rnd := rand.New(rand.NewSource(1024))

	for i := 0; ; i++ {
		rnd.Read(k)
		rnd.Read(v)
		if err := tree.Insert(cp(k), cp(v), nil); err != nil {
			panic(err)
		}
		if i%5300 == 0 {
			point := tree.ComputeCommitment().Bytes()
			fmt.Printf("Wrote %d elements to tree, in %v, speed %.02f items/ms, root %x\n",
				i, time.Since(start),
				float64(i*int(time.Millisecond))/float64(time.Since(start)),
				point)
		}
		if time.Since(start) > 10*time.Second {
			fmt.Printf("Wrote %d elements to tree, in %v\n", i, time.Since(start))
			break
		}
	}
}
