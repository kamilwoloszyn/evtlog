# evtlog

A small Go library for reading Windows Event Log channels using cgo and the Windows Event Log API (`wevtapi`).

This project is intended to be imported as a library and used by other Go applications on Windows.

## Features

- Reads events from a specific Windows event log channel
- Returns a Go-friendly result struct
- Supports bookmark-based pagination
- Supports optional Windows Event Log XML query filters
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
    // Read a limited, non-paginated result.
    result := evtlog.ReadChannelEvents("System", 10)
    if result.Error != 0 {
        fmt.Printf("error: %d\n", result.Error)
        return
    }

    for i, log := range result.Logs {
        fmt.Printf("Event #%d\n%s\n", i+1, log)
    }

    // Optional: provide an Event Log XML query filter.
    filtered := evtlog.ReadEvents("System", 10, `*[System[(EventID=7036)]]`)
    fmt.Printf("filtered count: %d, error: %d\n", filtered.Count, filtered.Error)
}
```

### Pagination

Pass an empty bookmark for the first page. Save `Bookmark` and pass it back to retrieve the next page. Continue while `HasMore` is true.

```go
bookmark := ""

for {
    page := evtlog.ReadEventsPage(
        "System",
        25,
        bookmark,
        `*[System[(EventID=7036)]]`, // optional filter
    )

    if page.Error != 0 {
        fmt.Printf("error: %d\n", page.Error)
        break
    }

    fmt.Printf("page count: %d\n", page.Count)
    for _, log := range page.Logs {
        fmt.Println(log)
    }

    if !page.HasMore {
        break
    }

    bookmark = page.Bookmark
}
```

For pagination without a filter, use:

```go
page := evtlog.ReadChannelEventsPaginated("System", bookmark, 25)
```

## API

```go
type EventLogResult struct {
    Logs     []string
    Count    int
    Error    uint32
    Bookmark string
    HasMore  bool
}

func ReadChannelEvents(channelPath string, maxEvents uint32) EventLogResult
func ReadEvents(channelPath string, maxEvents uint32, eventQuery ...string) EventLogResult
func ReadEventsPage(channelPath string, maxEvents uint32, bookmark string, eventQuery ...string) EventLogResult
func ReadChannelEventsPaginated(channelPath, bookmark string, maxEvents uint32) EventLogResult
```

The `eventQuery` argument is optional. It is passed directly to the Windows Event Log XML query engine.

## Notes

- `ReadChannelEvents` and `ReadEvents` read up to `maxEvents` entries and are still available for non-paginated use.
- Pagination uses an opaque Windows Event Log bookmark, not a numeric offset.
- `Bookmark` identifies the last event returned by a page.
- `HasMore` indicates that another event can be read after the current page.
- A bookmark should be used with the same channel and event query.
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
