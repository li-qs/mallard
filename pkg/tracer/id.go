package tracer

import (
	"fmt"
	"time"
)

func GenerateTraceID() string {
	return fmt.Sprintf("%x%x", time.Now().UnixNano(), time.Now().UnixNano()%1000000)
}
