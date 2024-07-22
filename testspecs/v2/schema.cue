_#Test: {
	name:         string
	interpreter?: string
	workdir?:     string

	before_commands?: [...string]
	after_commands?: [...string]
	commands?: [...string]

	tests?: [..._#Test]
}

_#Resolved: {
	name:        string
	interpreter: string
	workdir:     string

	before_commands: [...string]
	after_commands: [...string]
	commands: [...string]

	tests: [..._#Resolved]
}

_defaults: _#Resolved & {
	name:        ""
	interpreter: "/bin/sh"
	workdir:     "."

	before_commands: []
	after_commands: []
	commands: []

	tests: []
}

_#Resolver: {
	parent:         _#Resolved
	this_test=test: _#Test

	resolved: _#Resolved & {
		// name: "\(parent.name)/\(self.name)"
		name: parent.name
		// interpreter: *self.interpreter | parent.interpreter
		// workdir:     *self.workdir | parent.workdir
		// commands:    *self.commands | parent.commands

		// before_commands: *(parent.before_commands + self.before_commands) | parent.before_commands

		tests: [
			for t in this_test.tests {
				(_#Resolver & {
					parent: _defaults
					test:   t
				}).resolved
			},
		]
	}
}

all_tests: {[string]: _#Test}
all_tests: {
	minimum: {name: "Minimum"}

	tree: {
		name: "Tree"
		tests: [{
			name: "Left"
			tests: [
				{name: "Left Leaf"},
				{name: "Right Leaf"},
			]
		}, {
			name: "Right"
			tests: [
				{name: "Left Leaf"},
				{name: "Right Leaf"},
			]
		}]
	}

	let build_cmd = "go build -o ./bin/smoke"

	full: {
		name:        "Smoke"
		interpreter: "/bin/sh"
		tests: [{
			name: "Builds"
			commands: [build_cmd]
		}, {
			name:    "Self-Tests"
			workdir: "testbeds"
			before_commands: [build_cmd]
			tests: [{
				name: "A"
			}, {
				name: "B"
			}]
		}]
	} // full
}

all_outputs: {[string]: _#Resolved}
all_outputs: {
	for name, t in all_tests {
		"\(name)": (_#Resolver & {
			parent: _defaults
			test:   t
		}).resolved
	}
}
