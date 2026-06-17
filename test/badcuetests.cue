// A typo'd field (`chekcs` for `checks`) on an otherwise-valid node. The cue
// schema must reject it as a closed-struct violation and exit 65 (EX_DATAERR),
// rather than silently dropping the field at Decode.
commands: ["echo hi"]
chekcs: ["stdout"]
