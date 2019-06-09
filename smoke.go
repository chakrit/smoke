package main

type (
	Command string

	Output struct {
		ExitCode int
		Stdout   []byte
		Stderr   []byte
	}

	TestResult struct {
		Test            *Test
		PreviousOutputs []*Output
		Outputs         []*Output
		Subresults      []*TestResult
	}
)
