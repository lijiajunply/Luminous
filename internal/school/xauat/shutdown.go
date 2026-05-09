package xauat

// Shutdown stops all cache cleanup goroutines. Call during server shutdown.
func Shutdown() {
	busCache.Stop()
	scoreCache.Stop()
	examCache.Stop()
	courseCache.Stop()
	programCache.Stop()
	infoCache.Stop()
	semesterDateCache.Stop()
	paymentCache.Stop()
}
