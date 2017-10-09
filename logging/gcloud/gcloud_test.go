package gcloud_test

import (
	"testing"
	"time"

	"github.com/tomogoma/authms/logging"
)

func TestGCloud_log(t *testing.T) {
	tt := []struct {
		name  string
		entry logging.Entry
	}{
		{
			name: "all entry fields",
			entry: logging.Entry{
				Level:   "test-level",
				Time:    time.Now(),
				Payload: "a test log",
				Fields: map[string]interface{}{
					"field1": "field value one",
					"field2": "field value one",
				},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// TODO
			//gc := gcloud.SetProject("test project ID", "test logger name")
			//gc.Log(tc.entry)
		})
	}
}
