package pkg1

import (
	"log"
	"os"
)

func SomeFunction() {
	panic("error")           // want "panic call"
	log.Fatal("fatal error") // want "log.Fatal call"
	os.Exit(1)               // want "os.Exit call"
}
