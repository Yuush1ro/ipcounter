package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math/bits"
	"net"
	"os"
)

type BitSet struct {
	data []uint64
}

func NewBitSet() *BitSet {
	size := uint64(1) << 32
	words := size / 64
	return &BitSet{
		data: make([]uint64, words),
	}
}

func (b *BitSet) Set(n uint32) {
	idx := n / 64
	pos := n % 64
	b.data[idx] |= 1 << pos
}

func (b *BitSet) Count() uint64 {
	var total uint64
	for _, word := range b.data {
		total += uint64(bits.OnesCount64(word))
	}
	return total
}

func ipToUint32(ip net.IP) (uint32, error) {
	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("invalid IPv4: %v", ip)
	}
	return binary.BigEndian.Uint32(ip), nil
}

func processFile(filename string, bitset *BitSet) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		ip := net.ParseIP(line)
		if ip == nil {
			continue
		}
		val, err := ipToUint32(ip)
		if err == nil {
			bitset.Set(val)
		}
	}

	return scanner.Err()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <filename>")
		return
	}

	filename := os.Args[1]
	bitset := NewBitSet()

	if err := processFile(filename, bitset); err != nil {
		panic(err)
	}

	fmt.Printf("Unique IPs: %d\n", bitset.Count())
}
