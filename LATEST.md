# Breaking Changes

-  Function signature change: func Do(Settings, http.Request) Result became func Do (Settings, http.Request) (*Result, error)

# New Features

- Kurl CLI has a new argument `-pl` to print all latencies instead of default statistics to stdout.
- Kurl Go Package has a new `Settings.Warm` field, and Kurl CLI has a new argument `-warm`. These enable the execution of one http request during warmup, not measured in the result.