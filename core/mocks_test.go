package core

import "net/http"

// mockURLGetter - allows stubbing out any external http calls via the http.Get,
// urlfetch.Get or other methods that match the interface
type mockURLGetter struct {
	err    error
	body   string
	status int
}

func (u mockURLGetter) Get(url string) (*http.Response, error) {
	if u.err != nil {
		return nil, u.err
	}
	r := http.Response{Body: mockReadCloser{err: u.err, data: []byte(u.body)}}
	r.StatusCode = u.status
	return &r, nil
}

// mockReadCloser - Used within the mockURLGetter to stub out response data.
type mockReadCloser struct {
	err  error
	data []byte
}

func (m mockReadCloser) Read(data []byte) (int, error) {
	if m.err != nil {
		return 0, m.err
	}

	copy(data, m.data)
	return len(m.data), nil
}

func (m mockReadCloser) Close() error {
	return nil
}
