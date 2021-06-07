package main

import "fmt"

func main() {
	m := make(map[int]int)
	m[10] = 10
	fmt.Println(m)
	changeMap(m)
	fmt.Println(m)
}

func changeMap(m map[int]int) {
	m[10] = 20
}
