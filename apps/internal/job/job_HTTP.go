package job

import (
	"errors"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	clientPb "github.com/squzy/squzy_generated/generated/logger"
	"net/http"
	"strings"
	"time"
)

const (
	timeout = 5 * time.Second
	httpPort = 80
	httpsPort = 443
)

type jobHTTP struct {
	methodType string
	url        string
	headers    map[string]string
	statusCode int
}

type httpError struct {
	time        *timestamp.Timestamp
	code        clientPb.StatusCode
	description string
	location    string
}

var (
	wrongStatusError = errors.New("WRONG_STATUS_CODE")
)

func (e *httpError) GetLogData() *clientPb.Log {
	port := httpPort
	if strings.HasPrefix(e.location, "https") {
		port = httpsPort
	}
	return &clientPb.Log{
		Code:        e.code,
		Description: e.description,
		Meta: &clientPb.MetaData{
			Id:       uuid.New().String(),
			Location: e.location,
			Port:     int32(port),
			Time:     e.time,
			Type:     clientPb.Type_Http,
		},
	}
}

func NewHttpError(time *timestamp.Timestamp, code clientPb.StatusCode, description string, location string) CheckError {
	return &httpError{
		time:        time,
		code:        code,
		description: description,
		location:    location,
	}
}

func (j *jobHTTP) Do() CheckError {
	client := &http.Client{
		Timeout: timeout,
	}

	req, _ := http.NewRequest(j.methodType, j.url, nil)

	for name, val := range j.headers {
		req.Header.Set(name, val)
	}

	resp, err := client.Do(req)
	if err != nil {
		return NewHttpError(
			ptypes.TimestampNow(),
			clientPb.StatusCode_Error,
			err.Error(),
			j.url,
		)
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != j.statusCode {
		return NewHttpError(
			ptypes.TimestampNow(),
			clientPb.StatusCode_Error,
			wrongStatusError.Error(),
			j.url,
		)
	}

	return NewHttpError(
		ptypes.TimestampNow(),
		clientPb.StatusCode_OK,
		"",
		j.url,
	)
}

func NewJob(method, url string, headers map[string]string, status int) *jobHTTP {
	return &jobHTTP{
		methodType: method,
		url:        url,
		headers:    headers,
		statusCode: status,
	}
}
