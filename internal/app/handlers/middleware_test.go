package handlers

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type CustomResponseWriter struct {
	body       []byte
	statusCode int
	header     http.Header
}

func NewCustomResponseWriter() *CustomResponseWriter {
	return &CustomResponseWriter{header: http.Header{}}
}

func (w *CustomResponseWriter) Header() http.Header { return w.header }

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	w.body = b
	return 0, nil
}

func (w *CustomResponseWriter) WriteHeader(statusCode int) { w.statusCode = statusCode }

func Test_loggingResponseWriter_Write(t *testing.T) {
	type fields struct {
		ResponseWriter http.ResponseWriter
		responseData   *responseData
	}
	type args struct{ b []byte }
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "positive test #1",
			fields:  fields{ResponseWriter: NewCustomResponseWriter(), responseData: &responseData{status: 200, size: 1024}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &loggingResponseWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				responseData:   tt.fields.responseData,
			}
			got, err := r.Write(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("loggingResponseWriter.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("loggingResponseWriter.Write() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loggingResponseWriter_WriteHeader(t *testing.T) {
	type fields struct {
		ResponseWriter http.ResponseWriter
		responseData   *responseData
	}
	type args struct{ statusCode int }
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "positive test #1",
			fields: fields{ResponseWriter: NewCustomResponseWriter(), responseData: &responseData{status: 200, size: 1024}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &loggingResponseWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				responseData:   tt.fields.responseData,
			}
			r.WriteHeader(tt.args.statusCode)
		})
	}
}

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func TestWithLogging(t *testing.T) {
	type args struct {
		next http.Handler
	}
	tests := []struct {
		name string
		args args
		want http.Handler
	}{{name: "positive test #1", args: args{next: &Handler{}}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithLogging(tt.args.next)
			if reflect.TypeOf(got) == reflect.TypeOf((*http.Handler)(nil)).Elem() {
				t.Errorf("not handler")
			}
			h := WithLogging(&Handler{})
			go http.ListenAndServe(":8084", h)
			r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
			r.Header.Set("Content-Type", "text/plain")
			h.ServeHTTP(NewCustomResponseWriter(), r)
		})
	}
}
