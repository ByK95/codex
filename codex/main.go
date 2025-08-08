package main

import "C"

import "sync"

var (
    counter int
    mu      sync.Mutex
)

//export Increment
func Increment() {
    mu.Lock()
    counter++
    mu.Unlock()
}

//export GetCount
func GetCount() int {
    mu.Lock()
    defer mu.Unlock()
    return counter
}

func main() {}
