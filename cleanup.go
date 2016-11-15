package main

var cleanupHooks = make([]func(), 0)

func addCleanupHook(hook func()) {
	cleanupHooks = append(cleanupHooks, hook)
}

func runCleanupHooks() {
	for _, hook := range cleanupHooks {
		hook()
	}
}
