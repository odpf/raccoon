package worker

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"sync"
	"testing"
)

type mockTopicCreator struct {
	mock.Mock
}

func (m *mockTopicCreator) CreateTopics(ctx context.Context, ts []kafka.TopicSpecification, opts ...kafka.CreateTopicsAdminOption) ([]kafka.TopicResult, error) {
	args := m.Called(ctx, ts, opts)
	return args.Get(0).([]kafka.TopicResult), args.Error(1)
}

func TestRouter(t *testing.T) {
	t.Run("Should return topic according to format", func(t *testing.T) {
		tc := &mockTopicCreator{}
		router := Router{
			m:             &sync.Mutex{},
			topicsCreator: tc,
			format:        "prefix_%s_suffix",
			topics:        make(map[string]string),
		}

		tc.On("CreateTopics", mock.Anything, mock.Anything, mock.Anything).Return([]kafka.TopicResult{{}}, nil)
		topic, _ := router.getTopic("topic")
		assert.Equal(t, "prefix_topic_suffix", topic)
	})

	t.Run("Should only create topic when it doesn't exist yet", func(t *testing.T) {
		tc := &mockTopicCreator{}
		router := Router{
			m:             &sync.Mutex{},
			topicsCreator: tc,
			format:        "p_%s_s",
			topics:        make(map[string]string),
		}

		tc.On("CreateTopics", mock.Anything, mock.Anything, mock.Anything).Return([]kafka.TopicResult{{}}, nil).Once()
		router.getTopic("topic")
		topic, err := router.getTopic("topic")
		assert.NoError(t, err)
		assert.Equal(t, "p_topic_s", topic)
		tc.AssertExpectations(t)
	})

	t.Run("Should return error when topic cannot be created", func(t *testing.T) {
		tc := &mockTopicCreator{}
		router := Router{
			m:             &sync.Mutex{},
			topicsCreator: tc,
			format:        "p_%s_s",
			topics:        make(map[string]string),
		}

		tc.On("CreateTopics", mock.Anything, mock.Anything, mock.Anything).Return([]kafka.TopicResult{{}}, errors.New("error"))
		topic, err := router.getTopic("topic")
		assert.Equal(t, "", topic)
		assert.Error(t, err, topic)
	})

	t.Run("Should return error when topic cannot be created 2", func(t *testing.T) {
		tc := &mockTopicCreator{}
		router := Router{
			m:             &sync.Mutex{},
			topicsCreator: tc,
			format:        "p_%s_s",
			topics:        make(map[string]string),
		}

		tc.On("CreateTopics", mock.Anything, mock.Anything, mock.Anything).Return([]kafka.TopicResult{{
			Topic: "p_topic_s",
			Error: kafka.Error{},
		}}, errors.New("error"))
		topic, err := router.getTopic("topic")
		assert.Equal(t, "", topic)
		assert.Error(t, err, topic)
	})
}
