package gokafka

import (
	"encoding/binary"
)

/*
FetchResponse => [TopicName [Partition ErrorCode HighwaterMarkOffset MessageSetSize MessageSet]]
  TopicName => string
  Partition => int32
  ErrorCode => int16
  HighwaterMarkOffset => int64
  MessageSetSize => int32

Field					Description
HighwaterMarkOffset		The offset at the end of the log for this partition. This can be used by the client to determine how many messages behind the end of the log they are.
MessageSet				The message data fetched from this partition, in the format described above.
MessageSetSize			The size in bytes of the message set for this partition
Partition				The id of the partition this response is for.
TopicName				The name of the topic this response entry is for.
*/

type TopicData struct {
	Partition           int32
	ErrorCode           int16
	HighwaterMarkOffset int64
	MessageSetSize      int32
	MessageSet          MessageSet
}

type FetchResponse []struct {
	//CorrelationId int32
	TopicName  string
	TopicDatas []TopicData
}

func DecodeFetchResponse(payload []byte) (FetchResponse, error) {
	offset := uint64(0)

	//fetchResponse.CorrelationId = int32(binary.BigEndian.Uint32(payload[offset:]))
	//offset += 4

	topicDataCount := binary.BigEndian.Uint32(payload[offset:])
	offset += 4

	fetchResponse := make([]struct {
		TopicName  string
		TopicDatas []TopicData
	}, topicDataCount)

	for i := uint32(0); i < topicDataCount; i++ {
		topicNameLength := uint64(binary.BigEndian.Uint16(payload[offset:]))
		offset += 2
		fetchResponse[i].TopicName = string(payload[offset : offset+topicNameLength])
		offset += topicNameLength

		topicDataCount := binary.BigEndian.Uint32(payload[offset:])
		offset += 4
		fetchResponse[i].TopicDatas = make([]TopicData, topicDataCount)
		for j := uint32(0); j < topicDataCount; j++ {
			fetchResponse[i].TopicDatas[j].Partition = int32(binary.BigEndian.Uint32(payload[offset:]))
			offset += 4
			fetchResponse[i].TopicDatas[j].ErrorCode = int16(binary.BigEndian.Uint16(payload[offset:]))
			offset += 2
			fetchResponse[i].TopicDatas[j].HighwaterMarkOffset = int64(binary.BigEndian.Uint64(payload[offset:]))
			offset += 8
			fetchResponse[i].TopicDatas[j].MessageSetSize = int32(binary.BigEndian.Uint32(payload[offset:]))
			offset += 4
			fetchResponse[i].TopicDatas[j].MessageSet = make([]Message, fetchResponse[i].TopicDatas[j].MessageSetSize)
			for k := int32(0); k < fetchResponse[i].TopicDatas[j].MessageSetSize; k++ {
				fetchResponse[i].TopicDatas[j].MessageSet[k].Offset = int64(binary.BigEndian.Uint64(payload[offset:]))
				offset += 8
				fetchResponse[i].TopicDatas[j].MessageSet[k].MessageSize = int32(binary.BigEndian.Uint32(payload[offset:]))
				offset += 4
				fetchResponse[i].TopicDatas[j].MessageSet[k].Crc = binary.BigEndian.Uint32(payload[offset:])
				offset += 4
				fetchResponse[i].TopicDatas[j].MessageSet[k].MagicByte = int8(payload[offset])
				offset += 1
				fetchResponse[i].TopicDatas[j].MessageSet[k].Attributes = int8(payload[offset])
				offset += 1
				keyLength := int32(binary.BigEndian.Uint32(payload[offset:]))
				offset += 4
				if keyLength == -1 {
					fetchResponse[i].TopicDatas[j].MessageSet[k].Key = nil
				} else {
					fetchResponse[i].TopicDatas[j].MessageSet[k].Key = make([]byte, keyLength)
					copy(fetchResponse[i].TopicDatas[j].MessageSet[k].Key, payload[offset:offset+uint64(keyLength)])
					offset += uint64(keyLength)
				}

				valueLength := int32(binary.BigEndian.Uint32(payload[offset:]))
				offset += 4
				if valueLength == -1 {
					fetchResponse[i].TopicDatas[j].MessageSet[k].Value = nil
				} else {
					fetchResponse[i].TopicDatas[j].MessageSet[k].Value = make([]byte, valueLength)
					copy(fetchResponse[i].TopicDatas[j].MessageSet[k].Value, payload[offset:offset+uint64(valueLength)])
					offset += uint64(valueLength)
				}
				if offset == uint64(len(payload)) {
					break
				}
			}
		}
	}

	return fetchResponse, nil
}
