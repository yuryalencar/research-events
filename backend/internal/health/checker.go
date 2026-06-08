package health

import "context"

// Registry holds all registered health checkers and runs them on demand.
//
// The open/closed principle in practice: new dependencies register themselves here;
// the handler never changes. Add a checker in main.go — that's the only file to touch.
type Registry struct {
	checkers []Checker
}

func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a checker to the registry.
// Checkers are run in registration order.
func (r *Registry) Register(c Checker) {
	r.checkers = append(r.checkers, c)
}

// RunAll executes every registered checker and returns the results.
// The second return value is false if any single checker reports unhealthy.
func (r *Registry) RunAll(ctx context.Context) (map[string]CheckResult, bool) {
	results := make(map[string]CheckResult)
	allHealthy := true

	for _, c := range r.checkers {
		result := c.Check(ctx)
		results[c.Name()] = result
		if result.Status != StatusHealthy {
			allHealthy = false
		}
	}

	return results, allHealthy
}
