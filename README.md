# ring-buffer
A simple implementation of ring buffer to read a large log file and process it. Produces two files:
1. occurrences.txt - unique word and the number of occurrences in the log file
2. word_count.txt - line number and the number of words that line had in the log file (this file is in `APPEND` mode, so delete it if you run the program again)

# How to run
There are two ways to run the program. Both take the inputfile as an argument from the user:
```
make build
./ring_buffer -input=inputfile
```
OR
```
go run ./ring_buffer.go -input=inputfile
```
### Enable profiling
If you want to check the cpu profile afterwards using `pprof` tool, add the cpuprofile flag when you run the program (this will make the run very slow). for e.g.
```
./ring_buffer -input=inputfile -cpuprofile=cpu.prof
```

You can analyze the profile later and later generate a jpeg or a gif of the call stack using:
```
go tool pprof cpu.prof
```

### Unit tests
```
make cover
```

### Benchmarks
```
make benchmark
```

On my machine:
```
go test -bench=. -benchmem ./...
goos: darwin
goarch: amd64
pkg: github.com/kulkarnisamr/ring-buffer
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
BenchmarkWrite-16    	 6638828	       176.7 ns/op	      88 B/op	       2 allocs/op
BenchmarkRead-16     	49120402	        24.83 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/kulkarnisamr/ring-buffer	12.976s
```

# Performance
### Runtime
```
Hardware Overview:

  Model Name:	MacBook Pro
  Model Identifier:	MacBookPro16,1
  Processor Name:	8-Core Intel Core i9
  Processor Speed:	2.4 GHz
  Number of Processors:	1
  Total Number of Cores:	8
  L2 Cache (per Core):	256 KB
  L3 Cache:	16 MB
  Hyper-Threading Technology:	Enabled
  Memory:	32 GB
```
1. 1GB file: total time taken `66s`
2. 10GB file: total time taken `675s`

### Complexity
1. Time - enqueue and dequeue operations on the ring buffer are `O(1)`. I am not sorting the output for the line number and word counts file which is 3.5GB in size and that would significantly add to the overall compute time (complexity `O(N^logN)`).
2. Space - `O(n)` where n is the ring buffer size

### CPU profile
Profile for processing a 1GB file: ![profile001](https://github.com/kulkarnisamr/ring-buffer/assets/3310957/30473dd0-e014-48f4-9335-378bccb82f16)



# Approach
![ringbuffer-379101352](https://github.com/kulkarnisamr/ring-buffer/assets/3310957/1cb551cf-e17e-47f0-bbd6-ae1f70321612)

### Pros
1. Constant space allocation- The space allocated always stays constant and no matter how big the input file it is guaranteed that the allocated space (at ay point in time) is capped by the ring buffer's size. Go slices are very efficient since the underlying array never changes and all we are doing is slicing and reslicing as we enqueue and dequeue elements.
2. Single writer and multiple readers - This design allows for concurrent processing of the data once it is available (once the log file offset increment and a line is scanned). This is analogous to any event driven systems
3. The code will scale and run faster if there are more logical CPUs available at process start up. 

### Cons
1. Custom ring buffer implementation- I would have preferred to use an open source existing implementation which has been battle tested in other production environments but most of them had issue with the single producer and multiple consumers pattern that I wanted to implement.
2. Open source options tried:
   - containers/ring: I tried to use standard library's containers/ring but it did not work as I intended it to when it came to concurrent access and I ran short of time. Having said that, it should be very easy to replace the current buffer with containers/ring without changing too much of the existing code. Other open soure libraries like github.com/smallnest/ring
   - https://go.abhg.dev/container/ring/: This worked fine (PR: https://github.com/kulkarnisamr/ring-buffer/pull/3) but it took twice the processing time for a 10GB file compared to the current implementation so I dropped it.
  
### Further improvements
1. It would be worth while checking how to use [sync.Cond](https://pkg.go.dev/sync#Cond)https://pkg.go.dev/sync#Cond and it's `BroadCast` and `Signal` methods for waking up go-routines. From all the standard library's documentation I read, it sounds like the Go language authors still prefer if developers used channels for most synchronization use cases.


