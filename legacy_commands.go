package sml

// legacyReadOnlyCommands is preserved for pre-extraction SML views.
// New applications should prefer host-owned action identifiers instead of
// command attributes. The SML renderer never executes these commands.
var legacyReadOnlyCommands = map[string]bool{
	"help": true, "status": true, "whoami": true, "jobs": true,
	"lab": true, "lab methods": true, "lab ongoing": true, "lab history": true,
	"api list": true, "plugins list": true, "sessions": true, "version": true,
}
