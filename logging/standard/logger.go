package standard

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/tomogoma/authms/logging"
)

func init() {
	logging.SetEntryLoggerFunc(Log)
}

func Log(e logging.Entry) {
	fields := "{"
	for k, v := range e.Fields {
		vStr, _ := json.Marshal(v)
		fields = fmt.Sprintf("%s\"%s\": \"%s\", ", fields, k, vStr)
	}
	fields = strings.TrimSuffix(fields, ", ") + "}"
	log.Printf("%s - %s: %s %s", e.Level, e.Time, e.Payload, fields)
}
