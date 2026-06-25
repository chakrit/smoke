package cases

// A shared #Case schema plus reusable cases — the DRY lever a cue.mod import
// unlocks. A spec elsewhere in the module imports this package instead of
// re-declaring the scaffold (the lowfat-pantry use case, minimized).
#Case: {
	name: string
	commands: [...string]
}

Echo: #Case & {
	name: "Echo"
	commands: ["echo cue-module import works"]
}
