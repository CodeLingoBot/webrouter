package main

import (
	"fmt"
)

func main() {
	m := make(map[string]string)
	delete(m, "abc")
	fmt.Println("success")
}
