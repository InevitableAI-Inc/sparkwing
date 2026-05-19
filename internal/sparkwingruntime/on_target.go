package sparkwingruntime

import (
	"sort"

	"github.com/sparkwing-dev/sparkwing/sparkwing"
)

// EffectiveJobTargets returns the resolved target-set per JobNode id
// in the Plan. The result captures the author's OnTarget declarations
// plus the inferred propagation walk:
//
//  1. A node with an explicit OnTargets list uses that list verbatim.
//  2. A node without an explicit list whose consumers are all
//     non-universal inherits the union of their effective sets.
//  3. A node whose consumers include any universal set (or which
//     has no consumers and no explicit OnTarget) is universal.
//
// In the returned map, a nil/empty []string value marks a universal
// node. The set entries are sorted for stable comparison.
//
// Inferred propagation is computed in reverse-dependency order so
// downstream nodes (Job DAG leaves) decide upstream nodes (Job DAG
// roots) without iteration to a fixed point.
//
// EffectiveJobTargets is safe to call before dispatch; it does not
// mutate the Plan and ignores expansions that have not yet
// materialized.
func EffectiveJobTargets(p *sparkwing.Plan) map[string][]string {
	if p == nil {
		return nil
	}
	nodes := p.Nodes()
	if len(nodes) == 0 {
		return map[string][]string{}
	}

	consumers := make(map[string][]string, len(nodes))
	known := make(map[string]struct{}, len(nodes))
	for _, n := range nodes {
		known[n.ID()] = struct{}{}
	}
	for _, n := range nodes {
		for _, dep := range n.DepIDs() {
			if _, ok := known[dep]; !ok {
				continue
			}
			consumers[dep] = append(consumers[dep], n.ID())
		}
	}

	byID := make(map[string]*sparkwing.JobNode, len(nodes))
	for _, n := range nodes {
		byID[n.ID()] = n
	}

	indeg := make(map[string]int, len(nodes))
	for id := range known {
		indeg[id] = len(consumers[id])
	}
	queue := make([]string, 0, len(nodes))
	for id, d := range indeg {
		if d == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	const universal = "__universal__"
	effective := make(map[string]map[string]struct{}, len(nodes))

	processed := 0
	for len(queue) > 0 {
		var next []string
		for _, id := range queue {
			processed++
			n := byID[id]
			if explicit := n.OnTargets(); len(explicit) > 0 {
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

			for _, dep := range n.DepIDs() {
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

	out := make(map[string][]string, len(nodes))
	for _, n := range nodes {
		set, ok := effective[n.ID()]
		if !ok {
			// Cycle or unreachable node; treat as universal so we
			// never silently filter it out.
			out[n.ID()] = nil
			continue
		}
		if _, u := set[universal]; u {
			out[n.ID()] = nil
			continue
		}
		list := make([]string, 0, len(set))
		for t := range set {
			list = append(list, t)
		}
		sort.Strings(list)
		out[n.ID()] = list
	}
	return out
}

// JobAllowsTarget reports whether a job with the given effective
// target-set runs under the active target. An empty effective set
// (universal) matches every target including the empty selection;
// a non-empty effective set matches only when target is one of its
// entries. The empty active target intentionally fails any
// non-universal job -- a run without --for executes only the
// always-runs set.
func JobAllowsTarget(effective []string, target string) bool {
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

// EffectiveStepTargets is the WorkStep mirror of EffectiveJobTargets.
// It walks the Work's step DAG in reverse-dependency order and
// returns the resolved target set per step id. Spawn / SpawnEach
// items are ignored -- their target filtering happens at the spawned
// Job level when the dispatch fires.
func EffectiveStepTargets(w *sparkwing.Work) map[string][]string {
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
	byID := make(map[string]*sparkwing.WorkStep, len(steps))
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
