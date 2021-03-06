package healer

import (
	"encoding/json"
	"errors"
)

type BrokerConfig struct {
	ConnectTimeoutMS          int   `json:"connect.timeout.ms"`
	TimeoutMS                 int   `json:"timeout.ms"`
	TimeoutMSForEachAPI       []int `json:"timeout.ms.for.eachapi"`
	MetadataRefreshIntervalMS int   `json:"metadata.refresh.interval.ms"`
}

func DefaultBrokerConfig() *BrokerConfig {
	return &BrokerConfig{
		ConnectTimeoutMS:          60000,
		TimeoutMS:                 30000,
		TimeoutMSForEachAPI:       make([]int, 0),
		MetadataRefreshIntervalMS: 300 * 1000,
	}
}

func getBrokerConfigFromConsumerConfig(c *ConsumerConfig) *BrokerConfig {
	b := DefaultBrokerConfig()
	b.ConnectTimeoutMS = c.ConnectTimeoutMS
	b.TimeoutMS = c.TimeoutMS
	b.TimeoutMSForEachAPI = c.TimeoutMSForEachAPI
	return b
}

var (
	brokerAddressNotSet = errors.New("broker address not set in broker config")
)

func (c *BrokerConfig) checkValid() error {
	return nil
}

type ConsumerConfig struct {
	BootstrapServers     string `json:"bootstrap.servers"`
	ClientID             string `json:"client.id"`
	GroupID              string `json:"group.id"`
	RetryBackOffMS       int    `json:"retry.backoff.ms"`
	MetadataMaxAgeMS     int    `json:"metadata.max.age.ms"`
	SessionTimeoutMS     int32  `json:"session.timeout.ms"`
	FetchMaxWaitMS       int32  `json:"fetch.max.wait.ms"`
	FetchMaxBytes        int32  `json:"fetch.max.bytes"`
	FetchMinBytes        int32  `json:"fetch.min.bytes"`
	FromBeginning        bool   `json:"frombeginning"`
	AutoCommit           bool   `json:"auto.commit"`
	CommitAfterFetch     bool   `json:"commit.after.fetch"`
	AutoCommitIntervalMS int    `json:"auto.commit.interval.ms"`
	OffsetsStorage       int    `json:"offsets.storage"`
	ConnectTimeoutMS     int    `json:"connect.timeout.ms"`
	TimeoutMS            int    `json:"timeout.ms"`
	TimeoutMSForEachAPI  []int  `json:"timeout.ms.for.eachapi"`
}

func DefaultConsumerConfig() *ConsumerConfig {
	c := &ConsumerConfig{
		ClientID:             "",
		GroupID:              "",
		SessionTimeoutMS:     30000,
		RetryBackOffMS:       100,
		MetadataMaxAgeMS:     300000,
		FetchMaxWaitMS:       100,
		FetchMaxBytes:        10 * 1024 * 1024,
		FetchMinBytes:        1,
		FromBeginning:        false,
		AutoCommit:           true,
		CommitAfterFetch:     false,
		AutoCommitIntervalMS: 5000,
		OffsetsStorage:       1,
		ConnectTimeoutMS:     30000,
		TimeoutMS:            30000,
	}

	if c.TimeoutMSForEachAPI == nil {
		c.TimeoutMSForEachAPI = make([]int, 38)
		for i := range c.TimeoutMSForEachAPI {
			c.TimeoutMSForEachAPI[i] = c.TimeoutMS
		}
		c.TimeoutMSForEachAPI[API_JoinGroup] = int(c.SessionTimeoutMS) + 5000
		c.TimeoutMSForEachAPI[API_OffsetCommitRequest] = int(c.SessionTimeoutMS) / 2
		c.TimeoutMSForEachAPI[API_FetchRequest] = c.TimeoutMS + int(c.FetchMaxWaitMS)
	}

	return c
}

func GetConsumerConfig(config map[string]interface{}) (*ConsumerConfig, error) {
	b, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	c := DefaultConsumerConfig()
	err = json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	if c.TimeoutMSForEachAPI == nil {
		c.TimeoutMSForEachAPI = make([]int, 38)
		for i := range c.TimeoutMSForEachAPI {
			c.TimeoutMSForEachAPI[i] = c.TimeoutMS
		}
		c.TimeoutMSForEachAPI[API_JoinGroup] = int(c.SessionTimeoutMS) + 5000
		c.TimeoutMSForEachAPI[API_OffsetCommitRequest] = int(c.SessionTimeoutMS) / 2
		c.TimeoutMSForEachAPI[API_FetchRequest] = c.TimeoutMS + int(c.FetchMaxWaitMS)
	}

	return c, nil
}

var (
	emptyGroupID                 = errors.New("group.id is empty")
	invallidOffsetsStorageConfig = errors.New("offsets.storage must be 0 or 1")
)

func (config *ConsumerConfig) checkValid() error {
	if config.BootstrapServers == "" {
		return bootstrapServersNotSet
	}
	if config.GroupID == "" {
		return emptyGroupID
	}
	if config.OffsetsStorage != 0 && config.OffsetsStorage != 1 {
		return invallidOffsetsStorageConfig
	}
	return nil
}

type ProducerConfig struct {
	BootstrapServers         string `json:"bootstrap.servers"`
	ClientID                 string `json:"client.id"`
	Acks                     int16  `json:"acks"`
	CompressionType          string `json:"compress.type"`
	BatchSize                int    `json:"batch.size"`
	MessageMaxCount          int    `json:"message.max.count"`
	FlushIntervalMS          int    `json:"flush.interval.ms"`
	MetadataMaxAgeMS         int    `json:"metadata.max.age.ms"`
	FetchTopicMetaDataRetrys int    `json:"fetch.topic.metadata.retrys"`
	ConnectionsMaxIdleMS     int    `json:"connections.max.idle.ms"`

	// TODO
	Retries          int   `json:"retries"`
	RequestTimeoutMS int32 `json:"request.timeout.ms"`
}

func DefaultProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		ClientID:                 "healer",
		Acks:                     1,
		CompressionType:          "none",
		BatchSize:                16384,
		MessageMaxCount:          1024,
		FlushIntervalMS:          200,
		MetadataMaxAgeMS:         300000,
		FetchTopicMetaDataRetrys: 3,
		ConnectionsMaxIdleMS:     540000,

		Retries:          0,
		RequestTimeoutMS: 30000,
	}
}

var (
	messageMaxCountError   = errors.New("message.max.count must > 0")
	flushIntervalMSError   = errors.New("flush.interval.ms must > 0")
	unknownCompressionType = errors.New("unknown compression type")
	bootstrapServersNotSet = errors.New("bootstrap servers not set")
)

func (config *ProducerConfig) checkValid() error {
	if config.BootstrapServers == "" {
		return bootstrapServersNotSet
	}
	if config.MessageMaxCount <= 0 {
		return messageMaxCountError
	}
	if config.FlushIntervalMS <= 0 {
		return flushIntervalMSError
	}

	switch config.CompressionType {
	case "none":
	case "gzip":
	case "snappy":
	case "lz4":
	default:
		return unknownCompressionType
	}
	return nil
}
