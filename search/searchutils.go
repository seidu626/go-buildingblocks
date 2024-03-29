package searchutils

import (
	ahocorasick "github.com/seidu626/go-buildingblocks/datastructure"
	"sync"
)

// LinearSearchParallelInt is delegated to parallelize the execution of search method
func LinearSearchParallelInt(data []int, target, thread int) int {
	var (
		length      = len(data)
		dataXThread = length / thread
		oddment     = length % thread
		found       int
		wg          sync.WaitGroup
		result      []int
	)

	if oddment != 0 {
		oddment += thread * dataXThread
		found = LinearSearchInt(data[thread*dataXThread:oddment], target)

		if found != -1 {
			return found + thread*dataXThread
		} // else

		return -1
	}

	wg = sync.WaitGroup{}
	result = make([]int, thread)
	// var result []int
	wg.Add(thread)
	for i := 0; i < thread; i++ {
		go LinearSearchParallelIntHelper(&wg, data[i*dataXThread:(i+1)*dataXThread], target, i, result)
	}
	wg.Wait()
	//log.Println(result)
	for i := range result {
		if result[i] != -1 {
			return result[i] + i*dataXThread
		}
	}
	return -1
}

// LinearSearchParallelIntHelper is delegated to search the number and append to the given result array
func LinearSearchParallelIntHelper(wg *sync.WaitGroup, data []int, target, i int, result []int) {
	defer wg.Done()
	result[i] = LinearSearchInt(data, target)
}

// LinearSearchInt is a simple for delegated to find the target value
func LinearSearchInt(data []int, target int) int {
	var i int
	for i = range data {
		if target == data[i] {
			return i
		}
	}
	return -1
}

// ContainsString is delegated to verify if the given string is present in the data
func ContainsString(data, target string) bool {
	matcher := ahocorasick.NewStringMatcher([]string{target})
	return matcher.Contains([]byte(data))
}

// ContainsStringByte is delegated to verify if the given string is present in the data
func ContainsStringByte(data []byte, target string) bool {
	matcher := ahocorasick.NewStringMatcher([]string{target})
	return matcher.Contains(data)
}

// ContainsStrings is delegated to verify if the given array of string are present in the data
func ContainsStrings(data string, targets []string) bool {
	matcher := ahocorasick.NewStringMatcher(targets)
	return matcher.Contains([]byte(data))
}

// ContainsStringsByte is delegated to verify if the given array of string are present in the data
func ContainsStringsByte(data []byte, targets []string) bool {
	matcher := ahocorasick.NewStringMatcher(targets)
	return matcher.Contains(data)
}

// ContainsWhichStrings is delegated to verify which strings are present in the data
func ContainsWhichStrings(data string, target []string) []string {
	matcher := ahocorasick.NewStringMatcher(target)
	hits := matcher.Match([]byte(data))
	found := make([]string, len(hits))
	for i := range hits {
		found[i] = target[hits[i]]
	}
	return found
}

// ContainsWhichStringsByte is delegated to verify which strings are present in the data
func ContainsWhichStringsByte(data []byte, target []string) []string {
	matcher := ahocorasick.NewStringMatcher(target)
	hits := matcher.Match(data)
	found := make([]string, len(hits))
	for i := range hits {
		found[i] = target[hits[i]]
	}
	return found
}
