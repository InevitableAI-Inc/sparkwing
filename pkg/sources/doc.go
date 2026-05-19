// Package sources loads sources.yaml -- the file that names the
// secret/config backends a pipeline target can bind to. Each entry
// describes one backend (a remote controller's vault, the macOS
// keychain, a local dotenv file, or the process environment) and is
// referenced by name from pipelines.yaml's target.source field.
//
// # Source precedence (per-field, repo wins)
//
//  1. .sparkwing/sources.yaml         -- team-shared, checked in
//  2. ~/.config/sparkwing/sources.yaml -- per-user additions / overrides
//
// A name in both files merges per non-zero field with repo values
// winning. The file's `default:` key names the source used when a
// pipeline target doesn't bind to a named source explicitly; the
// runtime resolves the default per-call.
//
// # Loading
//
// [Load] reads one file; [Resolve] applies the repo / user
// precedence and returns one [Source] by name; [Names] lists every
// declared source. Type discriminators are exported as
// [TypeRemoteController], [TypeMacosKeychain], [TypeFile], and
// [TypeEnv].
//
// # Shape (yaml)
//
//	default: team-vault
//	sources:
//	  team-vault:
//	    type: remote-controller
//	    controller: shared        # profile name from profiles.yaml
//	  prod-vault:
//	    type: remote-controller
//	    controller: prod
//	  local-keychain:
//	    type: macos-keychain
//	    service: sparkwing-pi
//	  dotenv:
//	    type: file
//	    path: .sparkwing/secrets.local.env
//	  shell-env:
//	    type: env
//	    prefix: SW_
package sources
