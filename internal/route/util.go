package route

// Get the number of params that need to be allocated for this route.
func NumParams(r *Route) int {
	ct := 0
	ps := r.Parts()
	for _, p := range ps {
		if p.Parameter() != "" {
			ct++
		}
	}
	return ct
}
