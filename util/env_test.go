package util

import (
	"fmt"
)

func ExampleMessage() {
	fmt.Println(Message("test"))
	fmt.Println(Message(""))
	// Output:
	// File: env_test.go; Function: github.com/chuan-yun/esmeralda/util.ExampleMessage; Line: 8; Message: test
	// File: env_test.go; Function: github.com/chuan-yun/esmeralda/util.ExampleMessage; Line: 9; Message: ""
}
