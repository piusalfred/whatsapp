package whatsapp_test

import (
	"fmt"

	"github.com/piusalfred/whatsapp"
)

func ExampleIsCorrectAPIVersion() {
	fmt.Println(whatsapp.IsCorrectAPIVersion("v16.1"))    // true
	fmt.Println(whatsapp.IsCorrectAPIVersion("v21.0"))    // true
	fmt.Println(whatsapp.IsCorrectAPIVersion("v15.9"))    // false
	fmt.Println(whatsapp.IsCorrectAPIVersion("v100.0"))   // true
	fmt.Println(whatsapp.IsCorrectAPIVersion("v0.0"))     // false
	fmt.Println(whatsapp.IsCorrectAPIVersion("v0.hello")) // false
	fmt.Println(whatsapp.IsCorrectAPIVersion("vhi.1"))    // false
	fmt.Println(whatsapp.IsCorrectAPIVersion("v16.0"))    // true
	fmt.Println(whatsapp.IsCorrectAPIVersion("16.1"))     // false

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
