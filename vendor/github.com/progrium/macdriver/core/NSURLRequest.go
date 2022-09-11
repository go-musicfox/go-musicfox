//go:build darwin
// +build darwin

package core

type NSURLRequest struct {
	gen_NSURLRequest
}

func NSURLRequest_Init(url NSURL) NSURLRequest {
	return NSURLRequest_requestWithURL_(url)
}
