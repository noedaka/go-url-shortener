package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func saveHeapProfile(profileName string) error {
	dir := "profiles"

	resp, err := http.Get("http://localhost:8080/debug/pprof/heap")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	filename := filepath.Join(dir, profileName)
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Profile saved: %s\n", filename)
	return nil
}

func main() {
	time.Sleep(2 * time.Second)

	if err := saveHeapProfile("result.pprof"); err != nil {
		fmt.Printf("Error saving base profile: %v\n", err)
		os.Exit(1)
	}
}
