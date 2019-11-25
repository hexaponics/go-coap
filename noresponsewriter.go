package coap

import "context"

var (
	resp2XXCodes = []COAPCode{Created, Deleted, Valid, Changed, Content}
	resp4XXCodes = []COAPCode{BadRequest, Unauthorized, BadOption, Forbidden, NotFound, MethodNotAllowed, NotAcceptable, PreconditionFailed, RequestEntityTooLarge, UnsupportedMediaType}
	resp5XXCodes = []COAPCode{InternalServerError, NotImplemented, BadGateway, ServiceUnavailable, GatewayTimeout, ProxyingNotSupported}
)

func isSet(n uint32, pos uint32) bool {
	val := n & (1 << pos)
	return (val > 0)
}

func powerOfTwo(exponent uint32) uint32 {
	if exponent != 0 {
		return (2 * powerOfTwo(exponent-1))
	}
	return 1
}

func (w *noResponseWriter) decodeNoResponseOption(v uint32) []COAPCode {
	var codes []COAPCode
	if v == 0 {
		// No suppresed code
		return codes
	}

	var i uint32
	// Max bit value:4; ref:table_2_rfc7967
	for i = 0; i <= 4; i++ {
		if isSet(v, i) {
			index := powerOfTwo(i)
			codes = append(codes, w.noResponseValueMap[index]...)
		}
	}
	return codes
}

type noResponseWriter struct {
	ResponseWriter
	noResponseValueMap map[uint32][]COAPCode
}

func newNoResponseWriter(w ResponseWriter) *noResponseWriter {
	return &noResponseWriter{
		ResponseWriter: w,
		noResponseValueMap: map[uint32][]COAPCode{
			2:  resp2XXCodes,
			8:  resp4XXCodes,
			16: resp5XXCodes,
		},
	}
}

func (w *noResponseWriter) Write(p []byte) (n int, err error) {
	return w.WriteWithContext(context.Background(), p)
}

func (w *noResponseWriter) WriteWithContext(ctx context.Context, p []byte) (n int, err error) {
	l, resp := prepareReponse(w, w.ResponseWriter.GetReq().Msg.Code(), w.ResponseWriter.GetCode(), w.ResponseWriter.GetContentFormat(), p)
	err = w.WriteMsgWithContext(ctx, resp)
	return l, err
}

func (w *noResponseWriter) SetCode(code COAPCode) {
	w.ResponseWriter.SetCode(code)
}

func (w *noResponseWriter) SetContentFormat(contentFormat MediaType) {
	w.ResponseWriter.SetContentFormat(contentFormat)
}

func (w *noResponseWriter) NewResponse(code COAPCode) Message {
	return w.ResponseWriter.NewResponse(code)
}

func (w *noResponseWriter) WriteMsg(msg Message) error {
	return w.WriteMsgWithContext(context.Background(), msg)
}

func (w *noResponseWriter) WriteMsgWithContext(ctx context.Context, msg Message) error {
	noRespValue, ok := w.ResponseWriter.GetReq().Msg.Option(NoResponse).(uint32)
	if !ok {
		return ErrNotSupported
	}
	suppressedCodes := w.decodeNoResponseOption(noRespValue)

	for _, code := range suppressedCodes {
		if code == msg.Code() {
			return ErrMessageNotInterested
		}
	}
	return w.ResponseWriter.GetReq().Client.WriteMsgWithContext(ctx, msg)
}
