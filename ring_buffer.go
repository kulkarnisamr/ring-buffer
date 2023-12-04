package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	defaultBufferSize = 1024 * 1024
	occurrencesFile   = "occurrences_10gb.txt"
	wordCountFile     = "word_count_10gb.txt"
)

// RingBuffer represents a ring buffer
type RingBuffer struct {
	buffer   []*LogEntry
	capacity int
	size     int
	// readOffset and writeOffset are not used in this implementation to their
	// fullest effect
	// since we want to update in-memory maps and that code cannot be
	// idempotent especially with counting words, however with a persistent
	// store, we can use these offsets and have logic that is idempotent
	readOffset  int
	writeOffset int
	writeLock   sync.Mutex
	readLock    sync.Mutex
}

type LogEntry struct {
	entryIndex int64
	data       []byte
}

// LogProcessor represents a concurrent log processor
type LogProcessor struct {
	mu                sync.Mutex
	buffer            *RingBuffer
	occurrenceCounter map[string]int64
	wordCounter       map[int64]int64
	wordCounterFile   *os.File
	occurrenceFile    *os.File
}

func (l *LogEntry) identifierWithNumberOfWords() string {
	return fmt.Sprintf("%v %v\n", l.entryIndex, len(strings.Fields(string(l.data))))
}

func main() {
	startTime := time.Now().UTC()

	// Set the number of available CPU cores
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	wordCountFile, err := os.OpenFile(wordCountFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer wordCountFile.Close()

	out, err := os.Create(occurrencesFile)
	if err != nil {
		panic("could not create file")
	}
	defer out.Close()

	// Initialize the log processor
	logProcessor := LogProcessor{
		buffer: NewRingBuffer(4096),
		//&RingBuffer{buffer: make([]*LogEntry, 0, 4096), capacity: 4096, size: 0},
		occurrenceCounter: make(map[string]int64),
		wordCounter:       make(map[int64]int64),
	}

	// Channel to signal completion of goroutines
	done := make(chan struct{})

	// Channel to send log lines to processing goroutines
	logLines := make(chan struct{}, 4096)

	// Wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start the log processing goroutines
	for i := 0; i < numCPU; i++ {
		wg.Add(1)
		go logProcessor.processLogLines(logLines, &wg, wordCountFile)
	}

	// Start a goroutine to read the log file and send lines to the processing goroutines
	go func() {
		logProcessor.readLogFile(os.Args[1], logLines)
		close(logLines) // Close the channel when done reading the log file
	}()

	// Start a goroutine to wait for all processing goroutines to finish
	go func() {
		wg.Wait()
		close(done) // Close the done channel to signal completion
	}()

	// Wait for all goroutines to finish
	<-done

	// Write the unique words and their occurrences
	for word, count := range logProcessor.occurrenceCounter {
		str := fmt.Sprintf("%s %d\n", word, count)
		_, err := out.WriteString(str)
		if err != nil {
			panic(err)
		}
	}

	// Flush the word counts one last time
	if len(logProcessor.wordCounter) > 0 {
		logProcessor.flushOccurrenceCounts(wordCountFile)
	}

	endTime := time.Now().UTC().Sub(startTime)
	fmt.Printf("total time taken: %v\n", endTime.Seconds())
}

func (l *LogProcessor) readLogFile(filename string, logLines chan<- struct{}) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, defaultBufferSize), defaultBufferSize)
	var lineCount int64
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		l.buffer.enqueue(&LogEntry{
			data:       []byte(line),
			entryIndex: lineCount,
		})
		logLines <- struct{}{} // Send the signal to the processing goroutines
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}

func (l *LogProcessor) processLogLines(logLines chan struct{}, wg *sync.WaitGroup, wordCounterFile *os.File) {
	defer wg.Done()

	for range logLines {
		l.buffer.writeLock.Lock()
		node := l.buffer.dequeue()
		words := strings.Fields(string(node.data))
		l.mu.Lock()
		for _, word := range words {
			l.occurrenceCounter[strings.ToLower(word)]++
		}
		l.wordCounter[node.entryIndex] = int64(len(words))
		if l.buffer.isFull() {
			l.flushOccurrenceCounts(wordCounterFile)
		}
		l.mu.Unlock()
		l.buffer.writeLock.Unlock()
	}
}

// RingBuffer methods
func (l *LogProcessor) flushOccurrenceCounts(file *os.File) {
	w := bufio.NewWriterSize(file, defaultBufferSize)

	for k, v := range l.wordCounter {
		_, err := w.WriteString(fmt.Sprintf("%d %d\n", k, v))
		if err != nil {
			return
		}
	}

	l.wordCounter = make(map[int64]int64)
}

func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		buffer:   make([]*LogEntry, 0, capacity),
		capacity: capacity,
		size:     0,
	}
}

func (b *RingBuffer) isFull() bool {
	return b.size == b.capacity
}

func (b *RingBuffer) enqueue(item *LogEntry) {
	b.writeLock.Lock()
	defer b.writeLock.Unlock()

	// Buffer is not full, append the item
	b.buffer = append(b.buffer, item)
	b.size++

	// increment the write offset
	b.writeOffset++

	// reset the write offset if it has reached the capacity
	if b.capacity == b.writeOffset {
		b.writeOffset = 0
	}
}

func (b *RingBuffer) dequeue() *LogEntry {
	b.readLock.Lock()
	defer b.readLock.Unlock()

	if b.size == 0 {
		return nil
	}

	item := b.buffer[0]
	b.buffer = b.buffer[1:]
	b.size--

	// increment the read offset
	b.readOffset++

	// reset the read offset if it has reached the capacity
	if b.capacity == b.readOffset {
		b.readOffset = 0
	}

	return item
}
