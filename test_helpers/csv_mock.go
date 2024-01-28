package test_helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func SetupCSVOrder(dirPath string) {

	// Specify the filename
	filename := "example1.csv"

	// Create the full file path
	fullPath := filepath.Join(dirPath, filename)

	validRow := fmt.Sprintf("any_order_id;padberg_group;100.00;%s", time.Now().Format(time.DateOnly))
	validCsvContent := fmt.Sprintf("id;merchant_reference;amount;created_at\n%s", validRow)

	// Call the function to create the file
	err := createFile(fullPath, validCsvContent)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	const maxRetries = 10
	for i := 0; i < maxRetries; i++ {
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Println("File created successfully:", fullPath)
			return
		} else {
			fmt.Println("Waiting for file to be detected. Attempt:", i+1)
			time.Sleep(1 * time.Second)
		}
	}

	fmt.Println("File could not be detected after 10 retries.")
}

// createFile creates a file at the specified path.
func createFile(filePath, content string) error {
	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close() // Ensure the file is closed when done

	// Write content to the file
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func RemoveCSVOrder(path string) {
	_ = os.Remove(path + "example1.csv")
}
