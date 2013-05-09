package uwsgi

import (
	"strings"
	"net/http"
	"net/http/cgi"
)


//extern uwsgi_takeover
func uwsgi_takeover()
//extern uwsgi_response_write_body_do
func uwsgi_response_write_body_do(*interface{}, *byte, uint64) int
//extern uwsgi_response_prepare_headers_int
func uwsgi_response_prepare_headers_int(*interface{}, int) int
//extern uwsgi_response_add_header
func uwsgi_response_add_header(*interface{}, *byte, uint16, *byte, uint16) int
//extern uwsgi_request_body_read
func uwsgi_request_body_read(*interface{}, []byte, uint64) int

func Env(wsgi_req *interface{}) *map[string]string {
	var env map[string]string
        env = make(map[string]string)
	return &env
}

func EnvAdd(env *map[string]string, k *[65536]byte, kl uint16, v *[65536]byte, vl uint16) {
	(*env)[ string((*k)[0:kl]) ] = string((*v)[0:vl])
}

type ResponseWriter struct {
        r       *http.Request
        wsgi_req *interface{}
        headers      http.Header
        wroteHeader bool
}

func (w *ResponseWriter) Write(p []byte) (n int, err error) {
        if !w.wroteHeader {
                w.WriteHeader(http.StatusOK)
        }
        uwsgi_response_write_body_do(w.wsgi_req, &p[0], uint64(len(p)))
        return len(p), err
}

func (w *ResponseWriter) WriteHeader(status int) {
        uwsgi_response_prepare_headers_int(w.wsgi_req, status )
        if w.headers.Get("Content-Type") == "" {
                w.headers.Set("Content-Type", "text/html; charset=utf-8")
        }
        for k := range w.headers {
                for _, v := range w.headers[k] {
                        v = strings.NewReplacer("\n", " ", "\r", " ").Replace(v)
                        v = strings.TrimSpace(v)
			kb := []byte(k)
			vb := []byte(v)
                        uwsgi_response_add_header(w.wsgi_req, &kb[0], uint16(len(k)), &vb[0], uint16(len(v)))
                }
        }
        w.wroteHeader = true
}

func (w *ResponseWriter) Header() http.Header {
        return w.headers
}


type BodyReader struct {
        wsgi_req *interface{}
}

// there is no close in request body
func (br *BodyReader) Close() error {
        return nil
}

func (br *BodyReader) Read(p []byte) (n int, err error) {
                return 0, err
}
/*
        m := len(p)
        var body []byte = uwsgi_request_body_read(br.wsgi_req, C.ssize_t(m), &rlen)
        if (body[0] == 0) {
                err = io.EOF;
                return 0, err
        } else if (body != nil) {
                return int(rlen), err
        }
        err = io.ErrUnexpectedEOF
        rlen = 0
        return int(rlen), err
}
*/

func RequestHandler(env *map[string]string, wsgi_req *interface{}) {
        httpReq, err := cgi.RequestFromMap(*env)
        if err == nil {
                httpReq.Body = &BodyReader{wsgi_req}
                w := ResponseWriter{httpReq, wsgi_req,http.Header{},false}
                http.DefaultServeMux.ServeHTTP(&w, httpReq)
        }
}


func Run() {
	uwsgi_takeover()
}