package column

func ColIndex(s string) int {
	result := 0

	// Iterate through each character in the column name
	for _, ch := range s {
		result = result*26 + int(ch-'A'+1) // Convert each letter to its corresponding number
	}

	// Adjust for zero-based indexing
	return result - 1
}

func ColFromIndex(n int) string {
	result := ""

	// Increment the number by 1 to adjust for one-based indexing
	n++

	// Convert the number to column name iteratively
	for n > 0 {
		remainder := (n - 1) % 26                     // Calculate the remainder
		result = string(rune('A'+remainder)) + result // Convert the remainder to character and prepend to result
		n = (n - 1) / 26                              // Update the number by dividing it by 26
	}

	return result
}
