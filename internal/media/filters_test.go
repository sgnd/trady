package media

// resetFilters forgets the cached CanBurnCaptions answer.
func resetFilters() {
	filtersMu.Lock()
	filters = nil
	filtersMu.Unlock()
}
