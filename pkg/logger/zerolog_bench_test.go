package logger

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func BenchmarkZerologLogger_With(b *testing.B) {
	b.ReportAllocs()
	logger, _ := NewZerologLogger(&Config{Output: ioutil.Discard})

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fields := benchmarkGenerateLogFields()
		b.StartTimer()
		logger.With(fields)
	}
}

// generates random structure for bench testing.
func benchmarkGenerateLogFields() Fields {
	min, max := 20, 30
	numFields := rand.Intn(max-min) + min

	fields := make(Fields, numFields)

	for i := 0; i < numFields; i++ {
		fields[fmt.Sprintf("test_%d", i)] = fmt.Sprintf("Test value for %d", i)
	}

	return fields
}
