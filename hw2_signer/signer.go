package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

//TH - count of threads to compute multi hash
const (
	TH = 6
)

func main() {
	//inputData := []int{0, 1, 1, 2, 3, 5, 8}
	inputData := []int{0, 1}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			fmt.Println("result:", dataRaw)
		}),
	}

	start := time.Now()
	ExecutePipeline(hashSignJobs...)
	end := time.Since(start)

	fmt.Println("Time", end)
}

//ExecutePipeline do
func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, elem := range jobs {
		wg.Add(1)
		out := make(chan interface{})
		go func(job job, in, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			defer close(out)
			job(in, out)
		}(elem, in, out, wg)
		in = out
	}

	wg.Wait()
}

//Calculate 2 operations as 1 secound instead linear 2 secound time
func crc32AndMd5Async(crc32 string, md5 string) (string, string) {
	dataChan := make(chan string)

	go func(dataChan chan string) {
		dataChan <- DataSignerCrc32(crc32)
	}(dataChan)

	md5 = DataSignerCrc32(md5)

	return <-dataChan, md5
}

//SingleHash doing some work
func SingleHash(in, out chan interface{}) {
	mutex := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for elem := range in {
		wg.Add(1)
		go singleHashWorker(elem, out, mutex, wg)
	}

	wg.Wait()
}

func singleHashWorker(in interface{}, out chan interface{}, mu *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	data := strconv.Itoa(in.(int))
	mu.Lock()
	md5 := DataSignerMd5(data)
	mu.Unlock()
	crc32, crc32md5 := crc32AndMd5Async(data, md5)
	out <- crc32 + "~" + crc32md5

	// fmt.Printf("%s SingleHash data %s\n", data, data)
	// fmt.Printf("%s SingleHash md5(data) %s\n", data, md5)
	// fmt.Printf("%s SingleHash crc32(md5(data)) %s\n", data, crc32md5)
	// fmt.Printf("%s SingleHash crc32(data) %s\n", data, crc32)
	// fmt.Printf("%s SingleHash result %s\n\n", data, result)
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for elem := range in {
		wg.Add(1)
		go multiHash(elem, out, wg)
	}

	wg.Wait()
}

func multiHash(in interface{}, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	data := in.(string)
	wgMulti := &sync.WaitGroup{}
	mutex := &sync.Mutex{}
	arr := make([]string, TH)

	for i := 0; i < TH; i++ {
		wgMulti.Add(1)

		go func(arr []string, i int, wg *sync.WaitGroup, mu *sync.Mutex) {
			defer wgMulti.Done()

			res := DataSignerCrc32(strconv.Itoa(i) + data)
			// fmt.Printf("%s MultiHash: crc32(th+step1)) %d %s\n", data, i, res)
			mu.Lock()
			arr[i] = res
			mu.Unlock()
		}(arr, i, wgMulti, mutex)
	}

	wgMulti.Wait()
	result := strings.Join(arr, "")
	// fmt.Printf("%s MultiHash result: %s\n\n", data, result)
	out <- result
}

func CombineResults(in, out chan interface{}) {
	res := make([]string, 0)

	for elem := range in {
		res = append(res, (elem.(string)))
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	result := strings.Join(res, "_")
	// fmt.Printf("CombineResults result: %s\n\n", result)
	out <- result
}
