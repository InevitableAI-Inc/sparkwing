package sparkwing

import "sort"

// formatOnTargetSkip returns the skip-reason string used when a job
// or step is filtered out by the OnTarget walk. The shape matches
// the WhenRunner skip message so dashboard renderers can treat the
// two uniformly.
func formatOnTargetSkip(effective []string, target string) string {
	if target == "" {
		return "OnTarget " + formatTargetSet(effective) + " not satisfied; no target selected"
	}
	return "OnTarget " + formatTargetSet(effective) + " does not include active target " + quoteTarget(target)
}

func formatTargetSet(effective []string) string {
	if len(effective) == 0 {
		return "[]"
	}
	out := "["
	for i, t := range effective {
		if i > 0 {
			out += " "
		}
		out += quoteTarget(t)
	}
	return out + "]"
}

func quoteTarget(t string) string { return "\"" + t + "\"" }

// jobAllowsTarget mirrors internal/sparkwingruntime.JobAllowsTarget.
// Kept here so RunWork can apply OnTarget filtering without importing
// the runtime package (the runtime imports sparkwing, not the other
// way around).
func jobAllowsTarget(effective []string, target string) bool {
	if len(effective) == 0 {
		return true
	}
	if target == "" {
		return false
	}
	for _, t := range effective {
		if t == target {
			return true
		}
	}
	return false
}

// effectiveStepTargets mirrors internal/sparkwingruntime.EffectiveStepTargets.
// Kept here so RunWork's OnTarget filter can run without a circular
// import on the runtime package.
func effectiveStepTargets(w *Work) map[string][]string {
	if w == nil {
		return nil
	}
	steps := w.Steps()
	if len(steps) == 0 {
		return map[string][]string{}
	}

	known := make(map[string]struct{}, len(steps))
	for _, s := range steps {
		known[s.ID()] = struct{}{}
	}
	consumers := make(map[string][]string, len(steps))
	for _, s := range steps {
		for _, dep := range s.DepIDs() {
			if _, ok := known[dep]; !ok {
				continue
			}
			consumers[dep] = append(consumers[dep], s.ID())
		}
	}
	byID := make(map[string]*WorkStep, len(steps))
	for _, s := range steps {
		byID[s.ID()] = s
	}

	indeg := make(map[string]int, len(steps))
	for id := range known {
		indeg[id] = len(consumers[id])
	}
	queue := make([]string, 0, len(steps))
	for id, d := range indeg {
		if d == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	const universal = "__universal__"
	effective := make(map[string]map[string]struct{}, len(steps))

	for len(queue) > 0 {
		var next []string
		for _, id := range queue {
			s := byID[id]
			if explicit := s.OnTargetList(); len(explicit) > 0 {
				set := make(map[string]struct{}, len(explicit))
				for _, t := range explicit {
					set[t] = struct{}{}
				}
				effective[id] = set
			} else {
				cons := consumers[id]
				if len(cons) == 0 {
					effective[id] = map[string]struct{}{universal: {}}
				} else {
					merged := map[string]struct{}{}
					sawUniversal := false
					for _, cid := range cons {
						cs := effective[cid]
						if _, u := cs[universal]; u {
							sawUniversal = true
							break
						}
						for k := range cs {
							merged[k] = struct{}{}
						}
					}
					if sawUniversal {
						effective[id] = map[string]struct{}{universal: {}}
					} else {
						effective[id] = merged
					}
				}
			}
			for _, dep := range s.DepIDs() {
				if _, ok := known[dep]; !ok {
					continue
				}
				indeg[dep]--
				if indeg[dep] == 0 {
					next = append(next, dep)
				}
			}
		}
		sort.Strings(next)
		queue = next
	}

	out := make(map[string][]string, len(steps))
	for _, s := range steps {
		set, ok := effective[s.ID()]
		if !ok {
			out[s.ID()] = nil
			continue
		}
		if _, u := set[universal]; u {
			out[s.ID()] = nil
			continue
		}
		list := make([]string, 0, len(set))
		for t := range set {
			list = append(list, t)
		}
		sort.Strings(list)
		out[s.ID()] = list
	}
	return out
}
