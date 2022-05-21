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
- The values are 100-1000 bytes, 

The back-end is a hash sponge. After each 1000 writes to the backend, it output 
the sponge state.`
	fmt.Println(desc)

	tree := verkle.New()
	start := time.Now()
	k := make([]byte, 32)
	//v := make([]byte, 1024)
	v := make([]byte, 32)
	rnd := rand.New(rand.NewSource(1024))
	for i := 0; ; i++ {
		rnd.Read(k)
		rnd.Read(v)
		//vLen := 100 + (rnd.Uint32() % 900)
		if err := tree.Insert(k, v, nil); err != nil {
			panic(err)
		}
		if i%5300 == 0 {
			//point := []byte{}
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
