config: {
	interpreter: "/bin/sh"
	timeout:     "5s"
}
checks: ["exitcode", "stdout"]
tests: [
	{
		name:     "Echo"
		commands: ["echo cue-driven smoke test"]
	},
	{
		name:     "Arithmetic"
		commands: ["echo $((6 * 7))"]
	},
]
