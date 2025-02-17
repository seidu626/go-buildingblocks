package fileutils

import (
	"strings"
	"testing"

	"github.com/seidu626/go-buildingblocks/helper"
)

const gogputils string = `../config_testdata/files/test9`
const codeFolder string = `../`
const dante string = `../config_testdata/files/dante.txt`

func TestCountLinesFile(t *testing.T) {
	file := `../config_testdata/files/test1.txt`
	lines, err := CountLines(file, "", -1)
	if err != nil || lines != 112 {
		t.Log(err)
		t.Fail()
	}

	_, err = CountLines(file+"test", "", -1)
	if err == nil {
		t.Log(err)
		t.Fail()
	}
}

func BenchmarkCountLinesFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CountLines(dante, "", -1)
		if err != nil {
			b.Fail()
		}
	}
}

func TestGetFileContentTypeKO(t *testing.T) {
	file := `../config_testdata/files/test.txt`
	_, err := GetFileContentType(file)

	if err == nil {
		t.Log("Error -> ", err)
		t.Fail()
	}
	t.Log(err)
}

func TestGetFileContentTypeTXT(t *testing.T) {
	file := `../config_testdata/files/test1.txt`
	fileType, err := GetFileContentType(file)

	if err != nil {
		t.Log("Error -> ", err)
		t.Fail()
	}
	if !strings.Contains(fileType, "text/plain") {
		t.Log(fileType)
		t.Fail()
	}
}

func TestGetFileContentTypePDF(t *testing.T) {
	file := `../config_testdata/files/test2.pdf`
	fileType, err := GetFileContentType(file)

	if err != nil {
		t.Log("Error -> ", err)
		t.Fail()
	}
	t.Log(fileType)
}

func TestGetFileContentTypeZIP(t *testing.T) {
	file := `../config_testdata/ziputils/test1.zip`
	fileType, err := GetFileContentType(file)

	if err != nil {
		t.Log("Error -> ", err)
		t.Fail()
	}
	if fileType != "application/zip" {
		t.Log(fileType)
		t.Fail()
	}
}

func TestGetFileContentTypeODT(t *testing.T) {
	file := `../config_testdata/files/test3.odt`
	fileType, err := GetFileContentType(file)

	if err != nil {
		t.Log("Error -> ", err)
		t.Fail()
	}
	if fileType != "application/odt" {
		t.Log(fileType)
		t.Fail()
	}
}

func TestGetFileContentTypeDOCX(t *testing.T) {
	file := `../config_testdata/files/test4.docx`
	fileType, err := GetFileContentType(file)

	if err != nil {
		t.Log("Error -> ", err)
		t.Fail()
	}
	if fileType != "application/docx" {
		t.Log(fileType)
		t.Fail()
	}
}

func TestGetFileContentTypeDOC(t *testing.T) {
	file := `../config_testdata/files/test5.doc`
	fileType, err := GetFileContentType(file)

	if err != nil {
		t.Log("Error -> ", err)
		t.Fail()
	}
	if fileType != "application/doc" {
		t.Log(fileType)
		t.Fail()
	}
}

func TestGetFileContentTypePickle(t *testing.T) {
	file := `../config_testdata/files/test6.pkl`
	fileType, err := GetFileContentType(file)

	if err != nil {
		t.Log("Error -> ", err)
		t.Fail()
	}
	if fileType != "application/pickle" {
		t.Log(fileType)
		t.Fail()
	}
}

func TestGetFileContentTypeMP4(t *testing.T) {
	file := `../config_testdata/files/test7.mp4`
	fileType, err := GetFileContentType(file)

	if err != nil {
		t.Log("Error -> ", err)
		t.Fail()
	}
	if fileType != "video/mp4" {
		t.Log(fileType)
		t.Fail()
	}

	file = `../config_testdata/files/test8.mp4`
	fileType, err = GetFileContentType(file)

	if err != nil {
		t.Log("Error -> ", err)
		t.Fail()
	}
	if fileType != "video/mp4" {
		t.Log(fileType)
		t.Fail()
	}
}

func TestGetFileContentTypeBIN(t *testing.T) {
	file := `../config_testdata/files/test9`
	fileType, err := GetFileContentType(file)
	if err != nil {
		t.Log("Error -> ", err)
		t.Fail()
	}
	if fileType != "elf/binary" {
		t.Log(fileType)
		t.Fail()
	}
}

func TestListFile(t *testing.T) {
	t.Log(ListFiles(codeFolder))
}

func BenchmarkListFile(t *testing.B) {
	for n := 0; n < t.N; n++ {
		ListFiles(codeFolder)
	}
}

func TestFindFilesSensitive(t *testing.T) {
	if len(FindFiles(codeFolder, `FindMe`, false)) != 2 {
		t.Fail()
	}
}

func TestFindFilesInsensitive(t *testing.T) {
	if len(FindFiles(codeFolder, `findme`, false)) != 2 {
		t.Fail()
	}
}

func BenchmarkFindFilesSensitive(t *testing.B) {
	for n := 0; n < t.N; n++ {
		FindFiles(codeFolder, `FindMe`, true)
	}
}
func BenchmarkFindFilesInsensitive(t *testing.B) {
	for n := 0; n < t.N; n++ {
		FindFiles(codeFolder, `findme`, true)
	}
}

func TestGetFileSize(t *testing.T) {
	size, err := GetFileSize(gogputils)
	if err != nil {
		t.Fail()
	}

	kbSize := size / 1024
	t.Log(helper.ByteCountIEC(size))
	t.Log(helper.ByteCountSI(size))
	t.Log(size)
	t.Log(kbSize, "K")
}

func TestExtractWordFromFile(t *testing.T) {
	ExtractWordFromFile(dante)
}

func TestCompareBinaryFile(t *testing.T) {
	type args struct {
		file1 string
		file2 string
		nByte int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{

		{
			name: "ko",
			args: args{
				file1: gogputils,
				file2: dante,
				nByte: 0,
			},
			want: false,
		},
		{
			name: "ok",
			args: args{
				file1: dante,
				file2: dante,
				nByte: 0,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareBinaryFile(tt.args.file1, tt.args.file2, tt.args.nByte); got != tt.want {
				t.Errorf("CompareBinaryFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
