// Package runners loads runners.yaml -- the file that names the
// runners a pipeline can dispatch jobs to. Each entry declares the
// labels it advertises and, for cluster-backed types, the spec used
// to materialize a runner pod. Job-level selection (Job.Requires /
// Prefers / WhenRunner) matches against these advertised labels.
//
// # Source precedence (per-field, repo wins)
//
//  1. .sparkwing/runners.yaml         -- team-shared, checked in
//  2. ~/.config/sparkwing/runners.yaml -- per-user additions / overrides
//
// A name present in both files is merged with repo values winning
// per non-zero field; user-only fields fill blanks. Names only in
// the user file resolve as-is.
//
// Implicit local: if neither file declares a runner named "local",
// [Resolve]("local") and [Names] synthesize one carrying labels for
// the current host's OS and architecture. A user-declared "local"
// entry overrides the synthesized version.
//
// # Loading
//
// [Load] reads one file; [Resolve] applies the repo / user
// precedence and the implicit-local synth; [Names] lists every name
// any layer declares. [File] is the on-disk shape; one entry is a
// [Runner] with optional [Spec] (Kubernetes-only pod placement +
// [Toleration]s + [Resources]).
//
// # Shape (yaml)
//
//	runners:
//	  local:
//	    type: local
//	    labels: [local, "os=darwin"]
//	  cloud-linux:
//	    type: kubernetes
//	    controller: shared
//	    labels: [cloud-linux, "os=linux"]
//	    spec:
//	      nodeSelector: { karpenter.sh/nodepool: general }
//	      resources:
//	        requests: { cpu: 2, memory: 4Gi }
//	  mac-mini:
//	    type: static
//	    labels: [mac-mini, "os=macos"]
package runners
