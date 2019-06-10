![](https://raw.githubusercontent.com/chakrit/smoke/master/smoke.jpg)

*WARNING:* Smoking is seriously bad for your health. Be [smoke-free][1].

# SMOKE

> Microsoft claims that after code reviews, "smoke testing is the most
> cost-effective method for identifying and fixing defects in software"
>
> <cite>[Wikipedia][0]</cite>

SMOKE is a CLI tool that helps you quickly and easily do smoke testing on your
software. It works on the simple assumption that **code that produces the same
observably correct output exhibits correct behavior.**

This, of course, is not 100% true all the time but you can get pretty far with
reliability with just by writing a small smoke testing file and using SMOKE to
help you.

*WARNING:* This is not a replacement for proper testing regiment. Use your own
judgement and discretion if you are dealing with mission-critical software.

# WORKFLOW

SMOKE assumes the following **1st-run** workflow:

1. Writes a new feature.
2. Writes a `tests.yml` file.
3. Runs `smoke tests.yml`.
4. Eyeballs 👀 the output.
5. If it looks correct, commits it with `smoke -c tests.yml`.

Later on subsequent runs:

1. Make changes.
2. Runs `smoke tests.yml`
3. Gets GREEN if the output doesn't change.

When making changes that should change the output:

1. Make changes.
2. Runs `smoke tests.yml`
3. Gets RED if the output actually changes.
4. Eyeballs 👀 the changes.
5. If it looks correct, commits it with `smoke -c tests.yml`.

Repeat the 1st-run workflow if the changes are expected and they are supposed to
be the new definition "correct".

Committing will produce a `.lock.yml` file, which should be checked into source
control so that other engineers can also run `smoke tests.yml` to check.

# FILE FORMAT

See smoke's own `tests.yml` for an example. The file is written with YAML.

A quick description follows:

```yaml
config:
  interpreter: /bin/sh       # changes interpreter, if using some shell-specific feature
  timeout: 5s                # sets all command's timeouts
checks:
  - exitcode                 # check and record command's exit code
  - stdout                   # check and record command's standard output
  - stderr                   # check and record command's standard error
tests:
  - name: Basics
    commands:
      - go build -o ./bin/smoke -v .
      - goimports -l *.go checks/*.go engine/*.go internal/p/*.go specs/*.go
      - ./bin/smoke
  - name: Self Tests
    commands:
      - go build -o ./bin/smoke -v .
      - ./bin/smoke --no-color --list tests.yml
  - name: Errors
    commands:
      - unknowncommand_asdfqwerzcxv # should error
  - name: Run Configs
    tests:                   # Subtests inherit parent's config and checks
      - name: WorkDir
        config:              # config can be re-defined in subtests
          workdir: ./internal/p
        commands:
          - ls
      - name: Env
        config:
          env:
            - "YEAS_ITS=Working"
        commands:
          - echo $YEAS_ITS
```

There is actually only one set of hash keys which can be (theoretically-)infinitely
nested. The root of the YAML document is actually a "root test" definition.
Multiple tests can be defined as children of this "root test".

```yaml
name: "The name of your Test, defaults to the filename"
config:
  interpreter: /bin/bash    # interpreter to interpret commands
  timeout: 5s               # timeouts if commands doesn't exit in 5s
  workdir: .                # changes command's start-up working dir
  env:                      # you can modify the environment that the command will run in
    - "PATH=./bin:/bin"
checks:                     # define checks to record and perform for the commands
  - exitcode                # checks and records command's status code on exit
  - stdout                  # checks and records command's entire standard output stream
  - stderr                  # checks and records command's entire standard error stream
commands:                   # define commands that constitutes the test
  - go install -v .
tests:
  - name: Subtest One
    config:
      workdir: ..           # overrides workdir config from parent
    commands:               # subtests inherits parent's command list
      - smoke               # runs after `go install -v .` finish
```

# LICENSE

MIT

[0]: https://en.wikipedia.org/wiki/Smoke_testing_(software)
[1]: https://smokefree.gov/
