//go:build windows
// +build windows

package evtlog

/*
#cgo windows LDFLAGS: -lwevtapi
#include <windows.h>
#include <winevt.h>
#include <stdlib.h>

typedef struct {
    char **logs;
    size_t count;
    DWORD error;
} EventLogResult;

static void freeEventLogResult(EventLogResult *result) {
    if (result == NULL) {
        return;
    }

    if (result->logs != NULL) {
        for (size_t i = 0; i < result->count; ++i) {
            if (result->logs[i] != NULL) {
                free(result->logs[i]);
            }
        }
        free(result->logs);
        result->logs = NULL;
    }

    result->count = 0;
    result->error = ERROR_SUCCESS;
}

static void wideToUTF8(const wchar_t *input, char **output) {
    if (output == NULL) {
        return;
    }

    *output = NULL;
    if (input == NULL) {
        return;
    }

    int size = WideCharToMultiByte(CP_UTF8, 0, input, -1, NULL, 0, NULL, NULL);
    if (size <= 0) {
        return;
    }

    char *buffer = (char *)malloc((size_t)size);
    if (buffer == NULL) {
        return;
    }

    if (WideCharToMultiByte(CP_UTF8, 0, input, -1, buffer, size, NULL, NULL) == 0) {
        free(buffer);
        buffer = NULL;
        return;
    }

    *output = buffer;
}

static EventLogResult readChannelEventsWithQuery(const wchar_t *channelPath, const wchar_t *eventQuery, DWORD maxEvents) {
    EventLogResult result = {0};
    result.error = ERROR_SUCCESS;

    if (channelPath == NULL || channelPath[0] == L'\0' || maxEvents == 0) {
        result.error = ERROR_INVALID_PARAMETER;
        return result;
    }

    EVT_HANDLE hResults = NULL;
    EVT_HANDLE hEvent = NULL;
    EVT_HANDLE events[1] = { NULL };
    ULONG returned = 0;
    DWORD count = 0;

    result.logs = (char **)calloc(maxEvents, sizeof(char *));
    if (result.logs == NULL) {
        result.error = ERROR_NOT_ENOUGH_MEMORY;
        return result;
    }

    hResults = EvtQuery(
        NULL,
        channelPath,
        eventQuery,
        EvtQueryChannelPath
    );

    if (hResults == NULL) {
        result.error = GetLastError();
        freeEventLogResult(&result);
        return result;
    }

    while (count < maxEvents) {
        events[0] = NULL;
        returned = 0;

        if (!EvtNext(hResults, 1, events, INFINITE, 0, &returned)) {
            DWORD error = GetLastError();
            if (error == ERROR_NO_MORE_ITEMS) {
                break;
            }

            result.error = error;
            EvtClose(hResults);
            freeEventLogResult(&result);
            return result;
        }

        if (returned == 0 || events[0] == NULL) {
            break;
        }

        hEvent = events[0];

        DWORD bufferSize = 0;
        DWORD propertyCount = 0;
        if (!EvtRender(NULL, hEvent, EvtRenderEventXml, 0, NULL, &bufferSize, &propertyCount)) {
            DWORD error = GetLastError();
            if (error != ERROR_INSUFFICIENT_BUFFER) {
                result.error = error;
                EvtClose(hEvent);
                EvtClose(hResults);
                freeEventLogResult(&result);
                return result;
            }
        }

        LPWSTR xmlW = (LPWSTR)malloc((size_t)bufferSize);
        if (xmlW == NULL) {
            result.error = ERROR_NOT_ENOUGH_MEMORY;
            EvtClose(hEvent);
            EvtClose(hResults);
            freeEventLogResult(&result);
            return result;
        }

        if (!EvtRender(NULL, hEvent, EvtRenderEventXml, bufferSize, xmlW, &bufferSize, &propertyCount)) {
            result.error = GetLastError();
            free(xmlW);
            EvtClose(hEvent);
            EvtClose(hResults);
            freeEventLogResult(&result);
            return result;
        }

        char *xmlUTF8 = NULL;
        wideToUTF8(xmlW, &xmlUTF8);
        free(xmlW);

        if (xmlUTF8 == NULL) {
            result.error = ERROR_NOT_ENOUGH_MEMORY;
            EvtClose(hEvent);
            EvtClose(hResults);
            freeEventLogResult(&result);
            return result;
        }

        result.logs[count] = xmlUTF8;
        EvtClose(hEvent);
        count++;
    }

    result.count = count;
    EvtClose(hResults);
    return result;
}
*/
import "C"

import (
	"syscall"
	"unsafe"
)

// EventLogResult holds the event XML strings returned by the Windows event log channel.
type EventLogResult struct {
	Logs  []string
	Count int
	Error uint32
}

// ReadChannelEvents reads up to maxEvents entries from a Windows event log channel.
// The channel is a channel path, such as: Microsoft-Windows-Kernel-PnP/Configuration.
func ReadChannelEvents(channelPath string, maxEvents uint32) EventLogResult {
	return ReadEvents(channelPath, maxEvents)
}

// ReadEvents reads up to maxEvents entries from a Windows event log channel.
// The optional eventQuery can be omitted, or you can provide a filter such as:
// *[System[(EventID=7036)]]
func ReadEvents(channelPath string, maxEvents uint32, eventQuery ...string) EventLogResult {
	query := ""
	if len(eventQuery) > 0 {
		query = eventQuery[0]
	}

	utf16Path, err := syscall.UTF16PtrFromString(channelPath)
	if err != nil || utf16Path == nil {
		return EventLogResult{Error: uint32(C.ERROR_INVALID_PARAMETER)}
	}

	var utf16Query *uint16
	if query != "" {
		utf16Query, err = syscall.UTF16PtrFromString(query)
		if err != nil || utf16Query == nil {
			return EventLogResult{Error: uint32(C.ERROR_INVALID_PARAMETER)}
		}
	}

	cResult := C.readChannelEventsWithQuery(
		(*C.wchar_t)(unsafe.Pointer(utf16Path)),
		(*C.wchar_t)(unsafe.Pointer(utf16Query)),
		C.DWORD(maxEvents),
	)

	result := EventLogResult{
		Count: int(cResult.count),
		Error: uint32(cResult.error),
	}

	if cResult.logs != nil && cResult.count > 0 {
		logs := make([]string, 0, int(cResult.count))
		slice := (*[1 << 30]*C.char)(unsafe.Pointer(cResult.logs))[:int(cResult.count):int(cResult.count)]
		for _, s := range slice {
			if s != nil {
				logs = append(logs, C.GoString(s))
			}
		}
		result.Logs = logs
	}

	C.freeEventLogResult(&cResult)
	return result
}
