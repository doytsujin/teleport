/*
Copyright 2021 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package protocol

import (
	"io"

	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"go.mongodb.org/mongo-driver/x/mongo/driver/wiremessage"

	"github.com/gravitational/trace"
)

// Message defines common interface for MongoDB wire protocol messages.
type Message interface {
	GetHeader() MessageHeader
}

// MessageHeader represents parsed MongoDB wire protocol message header:
//
// https://docs.mongodb.com/master/reference/mongodb-wire-protocol/#standard-message-header
type MessageHeader struct {
	MessageLength int32
	RequestID     int32
	ResponseTo    int32
	OpCode        wiremessage.OpCode
}

// MessageOpMsg represents parsed OP_MSG wire message:
//
// https://docs.mongodb.com/master/reference/mongodb-wire-protocol/#op-msg
type MessageOpMsg struct {
	Header     MessageHeader
	Flags      wiremessage.MsgFlag
	SequenceID string
	Documents  []bsoncore.Document
	Checksum   uint32
}

// GetHeader returns the wire message header.
func (m *MessageOpMsg) GetHeader() MessageHeader {
	return m.Header
}

// MessageUnknown represents a wire message we don't currently support.
type MessageUnknown struct {
	Header  MessageHeader
	Payload []byte
}

// GetHeader returns the wire message header.
func (m *MessageUnknown) GetHeader() MessageHeader {
	return m.Header
}

// ReadMessage reads the next MongoDB wire protocol message from the reader.
func ReadMessage(reader io.Reader) (Message, error) {
	header, payload, err := readMessage(reader)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	switch header.OpCode {
	case wiremessage.OpMsg:
		return readOpMsg(*header, payload)
	default:
		return &MessageUnknown{
			Header:  *header,
			Payload: payload,
		}, nil
	}
}

func readMessage(reader io.Reader) (*MessageHeader, []byte, error) {
	// First read message header which is 16 bytes.
	var header [16]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return nil, nil, trace.Wrap(err)
	}
	length, requestID, responseTo, opCode, _, ok := wiremessage.ReadHeader(header[:])
	if !ok {
		return nil, nil, trace.BadParameter("failed to read message header %v", header)
	}
	// Then read the entire message body.
	payload := make([]byte, length-16)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, nil, trace.Wrap(err)
	}
	return &MessageHeader{
		MessageLength: length,
		RequestID:     requestID,
		ResponseTo:    responseTo,
		OpCode:        opCode,
	}, payload, nil
}

func readOpMsg(header MessageHeader, payload []byte) (*MessageOpMsg, error) {
	flags, rem, ok := wiremessage.ReadMsgFlags(payload)
	if !ok {
		return nil, trace.BadParameter("failed to read OP_MSG flags %v", payload)
	}
	sectionType, rem, ok := wiremessage.ReadMsgSectionType(rem)
	if !ok {
		return nil, trace.BadParameter("failed to read OP_MSG section type %v", payload)
	}
	var documents []bsoncore.Document
	var seqID string
	switch sectionType {
	case wiremessage.SingleDocument:
		var doc bsoncore.Document
		doc, rem, ok = wiremessage.ReadMsgSectionSingleDocument(rem)
		if !ok {
			return nil, trace.BadParameter("failed to read OP_MSG section single document %v", payload)
		}
		documents = append(documents, doc)
	case wiremessage.DocumentSequence:
		seqID, documents, rem, ok = wiremessage.ReadMsgSectionDocumentSequence(rem)
		if !ok {
			return nil, trace.BadParameter("failed to read OP_MSG section document sequence %v", payload)
		}
	}
	var checksum uint32
	if flags&wiremessage.ChecksumPresent != 0 {
		checksum, _, ok = wiremessage.ReadMsgChecksum(rem)
		if !ok {
			return nil, trace.BadParameter("failed to read OP_MSG checksum %v", payload)
		}
	}
	return &MessageOpMsg{
		Header:     header,
		Flags:      flags,
		SequenceID: seqID,
		Documents:  documents,
		Checksum:   checksum,
	}, nil
}
