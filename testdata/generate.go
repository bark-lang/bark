//go:build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type User struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Age       int      `json:"age"`
	Active    bool     `json:"active"`
	Score     float64  `json:"score"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
}

func generateUser(id int) User {
	return User{
		ID:        id,
		Name:      fmt.Sprintf("User %d", id),
		Email:     fmt.Sprintf("user%d@example.com", id),
		Age:       20 + (id % 50),
		Active:    id%2 == 0,
		Score:     float64(id%100) + 0.5,
		Tags:      []string{"tag1", "tag2", "tag3"},
		CreatedAt: "2024-01-15T10:30:00Z",
	}
}

func generateFile(filename string, count int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(`{"users":[`)

	encoder := json.NewEncoder(file)
	for i := 0; i < count; i++ {
		if i > 0 {
			file.WriteString(",")
		}
		if err := encoder.Encode(generateUser(i)); err != nil {
			return err
		}
	}

	file.WriteString(`],"meta":{"total":`)
	fmt.Fprintf(file, "%d", count)
	file.WriteString(`,"generated":true}}`)

	return nil
}

func main() {
	outDir := flag.String("out", ".", "output directory")
	flag.Parse()

	sizes := map[string]int{
		"small.json":  100,      // ~15KB
		"medium.json": 10000,    // ~1.5MB
		"large.json":  100000,   // ~15MB
		"xlarge.json": 1000000,  // ~150MB
	}

	for name, count := range sizes {
		path := filepath.Join(*outDir, name)
		fmt.Printf("Generating %s (%d users)...\n", name, count)
		if err := generateFile(path, count); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", name, err)
			os.Exit(1)
		}

		info, _ := os.Stat(path)
		fmt.Printf("  Created %s (%.2f MB)\n", name, float64(info.Size())/(1024*1024))
	}

	fmt.Println("Done.")
}
