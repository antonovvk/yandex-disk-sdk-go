package yadisk

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// Token
	testValidToken = Token{
		AccessToken: os.Getenv("YANDEX_TOKEN"),
	}
	testAppFolder    = os.Getenv("YANDEX_DISK_APP_FOLDER")
	testInvalidToken = Token{
		AccessToken: "AQA0AA00qEYz00WXA7olo",
	}
	// Disk
	testYaDisk, _                 = NewYaDisk(context.Background(), nil, &testValidToken)
	testYaDiskWithInvalidToken, _ = NewYaDisk(context.Background(), nil, &testInvalidToken)
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func TestNewYaDisk(t *testing.T) {
	type args struct {
		ctx    context.Context
		token  *Token
		client *http.Client
	}
	tests := []struct {
		name    string
		args    args
		want    YaDisk
		wantErr bool
	}{
		{"success_test", args{context.Background(), &testValidToken, http.DefaultClient}, testYaDisk, false},
		{"error_test", args{context.Background(), nil, nil}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewYaDisk(tt.args.ctx, tt.args.client, tt.args.token)
			if tt.wantErr {
				assert.Error(t, err, "Expecting error")
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_yandexDisk_GetDisk(t *testing.T) {
	type args struct {
		fields []string
	}
	tests := []struct {
		name    string
		args    args
		yaDisk  YaDisk
		wantD   *Disk
		wantErr bool
	}{
		{"success_test", args{[]string{"is_paid"}}, testYaDisk, &Disk{IsPaid: false}, false},
		{"error_test", args{[]string{}}, testYaDiskWithInvalidToken, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotD, err := tt.yaDisk.GetDisk(tt.args.fields)
			if tt.wantErr {
				assert.Error(t, err, "Expecting error")
				return
			}
			gotD.IsPaid = false // We may be testing on paid account
			assert.Equal(t, tt.wantD, gotD)
		})
	}
}

func Test_yandexDisk_PerformUpload(t *testing.T) {
	fileName := randStringBytes(10)
	canon := createFile(fileName, rand.Intn(100)*1e4)
	defer removeFile(fileName)

	filePath := path.Join(testAppFolder, fileName)
	link, err := testYaDisk.GetResourceUploadLink(filePath, nil, true)
	require.NoError(t, err)

	pu, err := testYaDisk.PerformUpload(link, openFile(fileName))
	require.NoError(t, err)
	require.NotNil(t, pu)

	status, err := testYaDisk.GetOperationStatus(link.OperationID, nil)
	require.NoError(t, err)
	assert.Equal(t, "success", status.Status)

	dl, err := testYaDisk.GetResourceDownloadLink(filePath, nil)
	require.NoError(t, err)

	data, err := testYaDisk.PerformDownload(dl)
	require.NoError(t, err)
	assert.Equal(t, canon, data)
}

func Test_yandexDisk_PerformPartialUpload(t *testing.T) {
	fileName := randStringBytes(10) + "_partial"
	createFile(fileName, rand.Intn(100)*1e4)
	defer removeFile(fileName)
	link, err := testYaDisk.GetResourceUploadLink(path.Join(testAppFolder, fileName), nil, true)
	if err != nil {
		t.Errorf("yandexDisk.GetResourceUploadLink() error = %v", err.Error())
	}
	pu, err := testYaDisk.PerformPartialUpload(link, openFile(fileName), rand.Int63n(100)*1e3)
	require.NoError(t, err)
	require.NotNil(t, pu)

	status, err := testYaDisk.GetOperationStatus(link.OperationID, nil)
	require.NoError(t, err)
	assert.Equal(t, "success", status.Status)
}

func createFile(name string, size int) []byte {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}()
	data := randStringBytes(size)
	if _, err := f.WriteString(data); err != nil {
		panic(err)
	}
	return []byte(data)
}

func removeFile(name string) {
	err := os.Remove(name)
	if err != nil {
		panic(err)
	}
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func openFile(name string) (buffer *bytes.Buffer) {
	data, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := data.Close()
		if err != nil {
			panic(err)
		}
	}()
	reader := bufio.NewReader(data)
	buffer = bytes.NewBuffer(make([]byte, 0))
	part := make([]byte, 1024)
	for {
		var count int
		if count, err = reader.Read(part); err != nil {
			break
		}
		buffer.Write(part[:count])
	}
	if err != io.EOF {
		log.Fatal("Error Reading ", name, ": ", err)
	} else {
		err = nil
	}
	return
}
