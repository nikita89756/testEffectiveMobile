package customerrors

import "fmt"

var(
	ErrPersonNotFound = fmt.Errorf("Person not found")
	ErrNothingToDelete = fmt.Errorf("Person for deleting not found")
	ErrNothingToUpdate = fmt.Errorf("Person for updating not found")
	ErrKeyNotFound = fmt.Errorf("Key not found")
)