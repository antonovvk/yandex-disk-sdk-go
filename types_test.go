package yadisk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_responseInfo_setResponseInfo(t *testing.T) {
	type args struct {
		status     string
		statusCode int
	}
	tests := []struct {
		name string
		ri   *responseInfo
		args args
	}{
		{"success_test", new(responseInfo), args{status: "200 OK", statusCode: 200}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ri.setResponseInfo(tt.args.status, tt.args.statusCode)
		})
	}
}

func TestError_Error(t *testing.T) {
	err := new(Error)
	err.ErrorID = "customError"
	tests := []struct {
		name string
		e    *Error
		want string
	}{
		{"success_test", err, "customError"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.e.Error())
		})
	}
}

func TestPerformUpload_handleError(t *testing.T) {
	type args struct {
		ri responseInfo
	}
	tests := []struct {
		name    string
		pu      *PerformUpload
		args    args
		wantErr bool
	}{
		{"success_test", new(PerformUpload), args{responseInfo{"201 created", 201}}, false},
		{"success_test", new(PerformUpload), args{responseInfo{"413 payload too large", 413}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pu.handleError(tt.args.ri)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
