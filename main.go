package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math/bits"
	"net"
	"os"
	"time"
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

	var processedBytes int64
	var lastPrinted int64
	start := time.Now()

	for scanner.Scan() {
		line := scanner.Text()
		processedBytes += int64(len(line)) + 1

		ip := net.ParseIP(line)
		if ip != nil {
			val, err := ipToUint32(ip)
			if err == nil {
				bitset.Set(val)
			}
		}

		if processedBytes-lastPrinted > 10*1024*1024 {
			printProgress(processedBytes, totalSize, start)
			lastPrinted = processedBytes
		}
	}

	printProgress(totalSize, totalSize, start)
	fmt.Println()

	return scanner.Err()
}

func printProgress(current, total int64, start time.Time) {
	width := 50
	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))

	bar := "[" + string(repeat('#', filled)) + string(repeat('-', width-filled)) + "]"

	elapsed := time.Since(start)
	eta := time.Duration(0)
	if percent > 0 {
		eta = time.Duration(float64(elapsed)/percent - float64(elapsed))
	}

	fmt.Printf("\r%s %6.2f%% | Elapsed: %s | ETA: %s",
		bar, percent*100,
		formatDuration(elapsed),
		formatDuration(eta))
}

func formatDuration(d time.Duration) string {
	secs := int(d.Seconds())
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	return fmt.Sprintf("%dm %ds", secs/60, secs%60)
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

	fmt.Printf("\nUnique IPs: %d\n", bitset.Count())
}
