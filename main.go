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

	info, err := file.Stat()
	if err != nil {
		return err
	}
	totalSize := info.Size()

	scanner := bufio.NewScanner(file)

	var processBytes int64
	var lastPrinted int64

	for scanner.Scan() {
		line := scanner.Text()
		processBytes += int64(len(line)) + 1

		ip := net.ParseIP(line)
		if ip == nil {
			val, err := ipToUint32(ip)
			if err == nil {
				bitset.Set(val)
			}
		}

		if processBytes-lastPrinted > 10*1024*1024 {
			printProgress(processBytes, totalSize)
			lastPrinted = processBytes
		}
	}

	printProgress(totalSize, totalSize)
	fmt.Println()

	return scanner.Err()
}

func printProgress(current, total int64) {
	width := 50 // ширина прогресс-бара
	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))

	bar := "[" + string(repeat('#', filled)) + string(repeat('-', width-filled)) + "]"
	fmt.Printf("\r%s %6.2f%%", bar, percent*100)
}

func repeat(char rune, count int) []rune {
	res := make([]rune, count)
	for i := range res {
		res[i] = char
	}
	return res
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
