package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"rawboard/internal/database"
)

func main() {
	fmt.Println("🔍 Testing Valkey connection...")

	uri := os.Getenv("VALKEY_URI")
	if uri == "" {
		uri = "redis://localhost:6379"
	}
	fmt.Printf("📡 Connecting to: %s\n", uri)

	db, err := database.NewValkeyDB()
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("✅ Connection established!")

	ctx := context.Background()

	// Test ping
	fmt.Print("🏓 Testing ping... ")
	if err := db.Ping(ctx); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ OK")

	// Test set operation
	testKey := "test:connection"
	testValue := "hello-valkey"
	fmt.Printf("📝 Testing SET %s=%s... ", testKey, testValue)
	if err := db.Set(ctx, testKey, testValue); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ OK")

	// Test get operation
	fmt.Printf("📖 Testing GET %s... ", testKey)
	value, err := db.Get(ctx, testKey)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	if value != testValue {
		fmt.Printf("❌ Mismatch: expected '%s', got '%s'\n", testValue, value)
		os.Exit(1)
	}
	fmt.Printf("✅ OK (got: '%s')\n", value)

	// Test performance
	fmt.Print("⚡ Testing performance (10 operations)... ")
	start := time.Now()
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("perf:test:%d", i)
		if err := db.Set(ctx, key, i); err != nil {
			fmt.Printf("❌ Failed on operation %d: %v\n", i, err)
			os.Exit(1)
		}
		if _, err := db.Get(ctx, key); err != nil {
			fmt.Printf("❌ Failed reading operation %d: %v\n", i, err)
			os.Exit(1)
		}
	}
	duration := time.Since(start)
	fmt.Printf("✅ OK (%v)\n", duration)

	fmt.Println("\n🎉 All tests passed! Valkey connection is working perfectly.")
}
