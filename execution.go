package ci

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ericaro/ci/format"
	"time"
)

//execution is a tool to run any execution, and keep: information about it.
type execution struct {
	version    [20]byte      // sha1 of all sha1 when the build has started, or ended (if the execution should change it.)
	start, end time.Time     // keep track of when
	errcode    int           // execution error code
	result     *bytes.Buffer // console output
}

//Marshal converts execution state into a "format" message.
func (x *execution) Marshal() *format.Execution { return x.Status(true) }

//Status return a format.Execution status,
// withResult true will also serialize the buffer's content.
func (x *execution) Status(withResult bool) *format.Execution {
	version := fmt.Sprintf("%x", x.version)
	start, end := x.start.Unix(), x.end.Unix()
	code := int32(x.errcode)
	result := x.result.String()
	f := &format.Execution{
		Version: &version,
		Start:   &start,
		End:     &end,
		Errcode: &code,
	}
	if withResult {
		f.Result = &result
	}
	return f
}

//Unmarshal restore an execution instance from the format.Exception message.
func (x *execution) Unmarshal(f *format.Execution) error {

	b, err := hex.DecodeString(f.GetVersion())
	if err != nil {
		return err
	}
	if len(b) != 20 {
		return fmt.Errorf("invalid sha1 length %d instead of 20", len(b))
	}
	copy(x.version[:], b)

	x.start = time.Unix(f.GetStart(), 0)
	x.end = time.Unix(f.GetEnd(), 0)

	x.errcode = int(f.GetErrcode())
	x.result = bytes.NewBufferString(f.GetResult())

	return nil
}
