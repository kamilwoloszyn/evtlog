# evtlog

A small Go library for reading Windows Event Log channels using cgo and the Windows Event Log API (`wevtapi`).

This project is intended to be imported as a library and used by other Go applications on Windows.

## Features

- Reads events from a specific Windows event log channel
- Returns a Go-friendly result struct
- Works with channel paths like:
  - `Microsoft-Windows-Kernel-PnP/Configuration`
  - `Application`
  - `System`

## Requirements

- Windows operating system
- Go 1.22 or newer
- A Windows SDK with `wevtapi.lib`
- CGO enabled

## Installation

Add the module to your project:

```bash
go mod init example.com/myapp

go get github.com/kamilwoloszyn/evtlog
```

Or if you are working inside a local checkout:

```bash
go mod edit -replace github.com/kamilwoloszyn/evtlog=../evtlog
```

## Usage

```go
package main

import (
    "fmt"
    evtlog "github.com/kamilwoloszyn/evtlog"
)

func main() {
    // Read all available events from a channel
    result := evtlog.ReadChannelEvents("System", 10)

    if result.Error != 0 {
        fmt.Printf("error: %d\n", result.Error)
        return
    }

    fmt.Printf("count: %d\n", result.Count)
    for i, log := range result.Logs {
        fmt.Printf("Event #%d\n%s\n", i+1, log)
    }

    // Optional: provide an Event Log XML query filter
    filtered := evtlog.ReadEvents("System", 10, `*[System[(EventID=7036)]]`)
    if filtered.Error != 0 {
        fmt.Printf("filtered error: %d\n", filtered.Error)
        return
    }

    fmt.Printf("filtered count: %d\n", filtered.Count)
}
```

## API

```go
type EventLogResult struct {
    Logs  []string
    Count int
    Error uint32
}

func ReadChannelEvents(channelPath string, maxEvents uint32) EventLogResult
func ReadEvents(channelPath string, maxEvents uint32, eventQuery ...string) EventLogResult
```

The `eventQuery` argument is optional and variadic. If you omit it, the library reads the channel without filtering. If you provide it, it is passed to the Windows Event Log XML query engine.

Example:

```go
// no filter
evtlog.ReadEvents("System", 10)

// filtered query
evtlog.ReadEvents("System", 10, `*[System[(EventID=7036)]]`)
```

## Notes

- `ReadChannelEvents` reads up to `maxEvents` entries from the requested channel.
- `ReadEvents` accepts the same arguments, with an optional XML event query filter as the last parameter.
- If the query is omitted, the function behaves like `ReadChannelEvents`.
- The returned logs are XML-formatted event strings from the Windows Event Log API.
- The library is Windows-only because it uses `winevt.h` and `wevtapi`.

## Example channel names

```go
"Application"
"System"
"Security"
"Microsoft-Windows-Kernel-PnP/Configuration"
"Microsoft-Windows-Win32k/Operational"
```

## Building

From a Windows machine with Go and the Windows SDK installed:

```bash
go build ./...
```

If you use a Windows environment without a default SDK path configured, make sure `wevtapi.lib` is available to the linker.
