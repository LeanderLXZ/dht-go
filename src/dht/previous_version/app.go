/**
* @Author: huadong.hu@outlook.com
* @Date: 7/7/20 8:51 PM
* @Desc:
 */

package main

//You Should not use time.sleep() to block go routines

import (
	"bufio"
	"log"
	"os"
	"strconv"
)

// Read a list of integers from `fileName`
// and launch `goRoutineNums` go routines to do sum up
// return sum of these Integers
func Sum(goRoutineNum int, fileName string) int {
	//TODO Add your code here
	if goRoutineNum < 1 {
		log.Println("goRoutineNum must >= 1!")
	}

	nums, err := readInts(fileName)
	if err != nil {
		log.Fatal(err)
	}

	// Create channel with size of go routine number
	ch := make(chan int)

	// Create go routines
	for i := 0; i < goRoutineNum; i++ {
		chunk := nums[i*len(nums)/goRoutineNum : (i+1)*len(nums)/goRoutineNum]
		go func() {
			sumRoutine := 0
			for _, v := range chunk {
				sumRoutine += v
			}
			ch <- sumRoutine
		}()
	}

	// Receive value from the channel and calculate the sum of go routines
	sum := 0
	for i := 0; i < goRoutineNum; i++ {
		sum += <-ch
	}
	return sum
}

//Read integers from reader
//Do not modify this function
func readInts(fileName string) ([]int, error) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	cin := bufio.NewScanner(file)
	cin.Split(bufio.ScanWords)
	var res []int
	for cin.Scan() {
		val, err := strconv.Atoi(cin.Text())
		if err != nil {
			return res, err
		}
		res = append(res, val)
	}
	return res, nil
}
