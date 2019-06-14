package tcglog

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type EventData interface {
	String() string
	RawBytes() []byte
	MeasuredBytes() []byte
}

type opaqueEventData struct {
	data          []byte
	informational bool
}

func (e *opaqueEventData) String() string {
	return ""
}

func (e *opaqueEventData) RawBytes() []byte {
	return e.data
}

func (e *opaqueEventData) MeasuredBytes() []byte {
	if !e.informational {
		return e.data
	}
	return nil
}

func bytesRead(stream *bytes.Reader) int {
	return int(stream.Size()) - stream.Len()
}

func makeEventDataImpl(pcrIndex PCRIndex, eventType EventType, data []byte,
	order binary.ByteOrder) (EventData, int, error) {
	switch {
	case pcrIndex == 8 || pcrIndex == 9:
		return makeEventDataGRUB(pcrIndex, eventType, data)
	default:
		return makeEventDataTCG(eventType, data, order)
	}
}

func makeOpaqueEventData(eventType EventType, data []byte) *opaqueEventData {
	switch eventType {
	case EventTypeEventTag, EventTypeSCRTMVersion, EventTypePlatformConfigFlags, EventTypeTableOfDevices,
		EventTypeNonhostInfo, EventTypeOmitBootDeviceEvents, EventTypeEFIGPTEvent:
		return &opaqueEventData{data: data, informational: false}
	default:
		return &opaqueEventData{data: data, informational: true}
	}
}

func makeEventData(pcrIndex PCRIndex, eventType EventType, data []byte,
	order binary.ByteOrder) (EventData, error) {
	event, n, err := makeEventDataImpl(pcrIndex, eventType, data, order)
	if event == nil {
		if err == io.EOF {
			err = errors.New("event data smaller than expected")
		}
		return makeOpaqueEventData(eventType, data), err
	}
	if n < len(data) {
		err = fmt.Errorf("event data contains %d bytes more than expected", len(data) - n)
	}
	return event, err
}
