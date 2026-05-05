package dotgen

const (
	// Alias represents a shell alias command.
	Alias = "alias"
	// Function represents a shell function command.
	Function = "function"
	// Raw represents raw shell code.
	Raw = "raw"
	// Run represents a command executed during generation.
	Run = "run"
)

// Kinds represents the supported command kinds.
//
//nolint:gochecknoglobals  // This is a constant list of supported kinds.
var Kinds = []string{Alias, Function, Raw, Run}
