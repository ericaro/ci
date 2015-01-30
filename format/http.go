package format

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
)

const (
	mimetype_pb = "application/x-protobuf"
)

//ProtoClient is a layer on top of an http.Client to handle just protobuf based
// exchange.
type ProtoClient struct {
	*http.Client
	URL string
}

//NewClient just create a default instance of ProtoClient.
func NewClient(url string) *ProtoClient {
	return &ProtoClient{
		Client: http.DefaultClient,
		URL:    url,
	}
}

//Proto is the new function on top of http.Client.
// performs a "POST" with a pbrequest encoded in the body, and wait for an http response with
// a format.Response  encoded.
func (c *ProtoClient) Proto(pbrequest *Request) (resp *Response, err error) {

	//create the request
	r, err := http.NewRequest("POST", c.URL, nil)
	if err != nil {
		return
	}

	//fill it with the proto Request object
	err = RequestEncode(r, pbrequest)
	if err != nil {
		return
	}

	//actually run the http layer
	httpr, err := c.Do(r)
	if err != nil {
		return
	}

	// and read the result.
	resp = new(Response)
	err = ResponseDecode(resp, httpr)
	if err != nil {
		return
	}
	return

}

//ResponseWriterEncode encode a format.Response into the ResponseWriter.
func ResponseWriterEncode(w http.ResponseWriter, data *Response) error {
	w.Header().Set("Content-Type", mimetype_pb)
	b, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	w.Write(b)
	return nil
}

//RequestDecode decodes 'r' the request into 'data'.
func RequestDecode(data *Request, r *http.Request) error {

	var err error
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return proto.Unmarshal(body, data)
}

//RequestEncode encodes 'data' into the request.
func RequestEncode(r *http.Request, data *Request) (err error) {
	var body bytes.Buffer
	b, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	_, err = body.Write(b)
	if err != nil {
		return err
	}
	// now exposes the buffer as a reader, for the client to read.
	rc := bytes.NewReader(body.Bytes())
	r.Body = ioutil.NopCloser(rc)
	return
}

func ResponseDecode(data *Response, r *http.Response) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return proto.Unmarshal(body, data)
}
