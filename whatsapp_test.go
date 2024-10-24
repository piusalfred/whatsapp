package whatsapp

import (
	"fmt"
)

// func main() {
//	fmt.Println(IsCorrectAPIVersion("v16.1"))  // true
//	fmt.Println(IsCorrectAPIVersion("v21.0"))  // true
//	fmt.Println(IsCorrectAPIVersion("v15.9"))  // false
//	fmt.Println(IsCorrectAPIVersion("v100.0")) // true
//}

func ExampleIsCorrectAPIVersion() {
	fmt.Println(IsCorrectAPIVersion("v16.1"))    // true
	fmt.Println(IsCorrectAPIVersion("v21.0"))    // true
	fmt.Println(IsCorrectAPIVersion("v15.9"))    // false
	fmt.Println(IsCorrectAPIVersion("v100.0"))   // true
	fmt.Println(IsCorrectAPIVersion("v0.0"))     // false
	fmt.Println(IsCorrectAPIVersion("v0.hello")) // false
	fmt.Println(IsCorrectAPIVersion("vhi.1"))    // false
	fmt.Println(IsCorrectAPIVersion("v16.0"))    // true
	fmt.Println(IsCorrectAPIVersion("16.1"))     // false

	// Output:
	// true
	// true
	// false
	// true
	// false
	// false
	// false
	// true
	// false
}
