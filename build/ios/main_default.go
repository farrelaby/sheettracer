//go:build !ios

package main

// main provides a buildable entrypoint for non-iOS platforms so `go build ./...`
// succeeds. The real iOS entrypoint is main_ios.go (//go:build ios).
func main() {}
