package byteutils

import (
	"bytes"
	"os"
	"testing"
)

const danteDataset string = `../config_testdata/files/dante.txt`

func BenchmarkTestIsUpperByteOK(b *testing.B) {
	content, err := os.ReadFile(danteDataset)
	if err != nil {
		return
	}
	content = bytes.ToUpper(content)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		IsUpperByte(content)
	}
}

func BenchmarkTestIsLowerByteKO(b *testing.B) {
	content, err := os.ReadFile(danteDataset)
	if err != nil {
		return
	}
	content = bytes.ToLower(content)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		IsLowerByte(content)
	}
}
