import "smoke.test/cuemod/cases"

config: {
	interpreter: "/bin/sh"
}
checks: ["exitcode", "stdout"]
tests: [cases.Echo]
