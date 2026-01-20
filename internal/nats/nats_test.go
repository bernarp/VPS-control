package nats

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventPayload_MarshalUnmarshal(t *testing.T) {
	t.Parallel()

	original := EventPayload[map[string]int]{
		ID:        "test-uuid-123",
		Timestamp: time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC),
		Version:   "1.0",
		Data:      map[string]int{"count": 42, "limit": 100},
	}

	bytes, err := json.Marshal(original)
	require.NoError(t, err)

	var restored EventPayload[map[string]int]
	err = json.Unmarshal(bytes, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Version, restored.Version)
	assert.True(t, original.Timestamp.Equal(restored.Timestamp))
	assert.Equal(t, original.Data["count"], restored.Data["count"])
	assert.Equal(t, original.Data["limit"], restored.Data["limit"])
}

type testCase interface {
	Name() string
	Run(t *testing.T)
}

type typedTestCase[T any] struct {
	name     string
	data     T
	validate func(
		t *testing.T,
		restored T,
	)
}

func (tc typedTestCase[T]) Name() string { return tc.name }

func (tc typedTestCase[T]) Run(t *testing.T) {
	original := EventPayload[T]{
		ID:        "id-123",
		Timestamp: time.Now().UTC(),
		Version:   "1.0",
		Data:      tc.data,
	}

	bytes, err := json.Marshal(original)
	require.NoError(t, err)

	var restored EventPayload[T]
	err = json.Unmarshal(bytes, &restored)
	require.NoError(t, err)

	tc.validate(t, restored.Data)
}

func TestEventPayload_DifferentTypes(t *testing.T) {
	t.Parallel()

	type UserEvent struct {
		UserID   int    `json:"user_id"`
		Username string `json:"username"`
		Active   bool   `json:"active"`
	}

	type MetricsEvent struct {
		CPU    float64  `json:"cpu"`
		Memory float64  `json:"memory"`
		Tags   []string `json:"tags"`
	}

	cases := []testCase{
		typedTestCase[string]{
			name: "string",
			data: "simple string data",
			validate: func(
				t *testing.T,
				restored string,
			) {
				assert.Equal(t, "simple string data", restored)
			},
		},

		typedTestCase[int]{
			name: "int",
			data: 42,
			validate: func(
				t *testing.T,
				restored int,
			) {
				assert.Equal(t, 42, restored)
			},
		},

		typedTestCase[[]string]{
			name: "slice",
			data: []string{"a", "b", "c"},
			validate: func(
				t *testing.T,
				restored []string,
			) {
				assert.Equal(t, []string{"a", "b", "c"}, restored)
			},
		},

		typedTestCase[UserEvent]{
			name: "struct_user",
			data: UserEvent{UserID: 1, Username: "alexey", Active: true},
			validate: func(
				t *testing.T,
				restored UserEvent,
			) {
				assert.Equal(t, 1, restored.UserID)
				assert.Equal(t, "alexey", restored.Username)
				assert.True(t, restored.Active)
			},
		},

		typedTestCase[MetricsEvent]{
			name: "struct_metrics",
			data: MetricsEvent{
				CPU:    75.5,
				Memory: 1024.0,
				Tags:   []string{"prod", "api"},
			},
			validate: func(
				t *testing.T,
				restored MetricsEvent,
			) {
				assert.InDelta(t, 75.5, restored.CPU, 0.001)
				assert.Equal(t, []string{"prod", "api"}, restored.Tags)
			},
		},

		typedTestCase[map[string]any]{
			name: "map_string_any",
			data: map[string]any{
				"string_field": "value",
				"number_field": float64(123),
				"bool_field":   true,
			},
			validate: func(
				t *testing.T,
				restored map[string]any,
			) {
				assert.Equal(t, "value", restored["string_field"])
				assert.Equal(t, float64(123), restored["number_field"])
				assert.Equal(t, true, restored["bool_field"])
			},
		},

		typedTestCase[*string]{
			name: "nil_pointer",
			data: nil,
			validate: func(
				t *testing.T,
				restored *string,
			) {
				assert.Nil(t, restored)
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(
			tc.Name(), func(t *testing.T) {
				t.Parallel()
				tc.Run(t)
			},
		)
	}
}

func TestEventPayload_JSONFormat(t *testing.T) {
	type SimpleData struct {
		Name string `json:"name"`
	}

	payload := EventPayload[SimpleData]{
		ID:        "uuid-abc",
		Timestamp: time.Date(2026, 1, 20, 15, 30, 0, 0, time.UTC),
		Version:   "2.0",
		Data:      SimpleData{Name: "test"},
	}

	bytes, err := json.Marshal(payload)
	require.NoError(t, err)

	var raw map[string]any
	err = json.Unmarshal(bytes, &raw)
	require.NoError(t, err)

	assert.Contains(t, raw, "id")
	assert.Contains(t, raw, "timestamp")
	assert.Contains(t, raw, "version")
	assert.Contains(t, raw, "data")

	assert.Equal(t, "uuid-abc", raw["id"])
	assert.Equal(t, "2.0", raw["version"])

	data, ok := raw["data"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "test", data["name"])
}

func TestEventPayload_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run(
		"empty_struct", func(t *testing.T) {
			type Empty struct{}

			payload := EventPayload[Empty]{
				ID:      "id",
				Version: "1.0",
				Data:    Empty{},
			}

			bytes, err := json.Marshal(payload)
			require.NoError(t, err)
			assert.Contains(t, string(bytes), `"data":{}`)
		},
	)

	t.Run(
		"unicode_data", func(t *testing.T) {
			payload := EventPayload[string]{
				ID:      "id",
				Version: "1.0",
				Data:    "Привет, 世界! 🚀",
			}

			bytes, err := json.Marshal(payload)
			require.NoError(t, err)

			var restored EventPayload[string]
			require.NoError(t, json.Unmarshal(bytes, &restored))
			assert.Equal(t, "Привет, 世界! 🚀", restored.Data)
		},
	)

	t.Run(
		"large_slice", func(t *testing.T) {
			bigSlice := make([]int, 10000)
			for i := range bigSlice {
				bigSlice[i] = i
			}

			payload := EventPayload[[]int]{
				ID:      "id",
				Version: "1.0",
				Data:    bigSlice,
			}

			bytes, err := json.Marshal(payload)
			require.NoError(t, err)

			var restored EventPayload[[]int]
			require.NoError(t, json.Unmarshal(bytes, &restored))
			assert.Len(t, restored.Data, 10000)
			assert.Equal(t, 9999, restored.Data[9999])
		},
	)

	t.Run(
		"zero_timestamp", func(t *testing.T) {
			payload := EventPayload[string]{
				ID:        "id",
				Timestamp: time.Time{},
				Version:   "1.0",
				Data:      "test",
			}

			bytes, err := json.Marshal(payload)
			require.NoError(t, err)

			var restored EventPayload[string]
			require.NoError(t, json.Unmarshal(bytes, &restored))
			assert.True(t, restored.Timestamp.IsZero())
		},
	)
}

func TestEventPayload_InvalidJSON(t *testing.T) {
	t.Parallel()

	t.Run(
		"malformed_json", func(t *testing.T) {
			var payload EventPayload[string]
			err := json.Unmarshal([]byte(`{invalid`), &payload)
			assert.Error(t, err)
		},
	)

	t.Run(
		"wrong_data_type", func(t *testing.T) {
			type Expected struct {
				Field int `json:"field"`
			}

			jsonStr := `{"id":"1","timestamp":"2026-01-20T00:00:00Z","version":"1.0","data":"not an object"}`

			var payload EventPayload[Expected]
			err := json.Unmarshal([]byte(jsonStr), &payload)
			assert.Error(t, err)
		},
	)

	t.Run(
		"missing_data_field", func(t *testing.T) {
			jsonStr := `{"id":"1","timestamp":"2026-01-20T00:00:00Z","version":"1.0"}`

			var payload EventPayload[string]
			err := json.Unmarshal([]byte(jsonStr), &payload)
			assert.NoError(t, err)
			assert.Empty(t, payload.Data)
		},
	)
}
