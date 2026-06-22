// Schema for `.cue` test specs. cueLoader unifies the user's file against
// `#Test` before decoding, so typo'd fields and wrong types fail closed as CUE
// constraint errors instead of being silently dropped at Decode. Closedness is
// recursive, so nested `tests` and `config` typos are caught too.

#Test: {
	name?:    string
	include?: string
	config?:  #Config
	commands?: [...string]
	checks?: [...string]
	tests?: [...#Test]
}

#Config: {
	workdir?: string
	env?: [...string]
	interpreter?: string
	timeout?:     string
}
