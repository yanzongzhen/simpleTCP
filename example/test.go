package main

import (
	"bytes"
	"fmt"
)

func main() {

	a := []byte{0, 10}
	b := []byte{0, 10}

	fmt.Println(bytes.Equal(a, b))

}
