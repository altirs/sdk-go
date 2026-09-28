package altirs

import "runtime/debug"

// sdkVersion is resolved from the build info Go records for this module, so
// the module tag stays the only place the version is declared.
var sdkVersion = resolveVersion()

func resolveVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	const modulePath = "github.com/altirs/sdk-go"
	for _, dep := range info.Deps {
		if dep.Path == modulePath {
			if dep.Version != "" {
				return dep.Version
			}
			break
		}
	}
	// Built from this module directly rather than consumed as a dependency.
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	return "dev"
}
