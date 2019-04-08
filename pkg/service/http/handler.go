package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/makerdao/testchain-deployment/pkg/serror"
	"github.com/makerdao/testchain-deployment/pkg/service/protocol"
	"github.com/sirupsen/logrus"
)

//HandlerMethod - method will be handled from Main http handler
type HandlerMethod func(log *logrus.Entry, ID string, requestBytes []byte) (response []byte, error *serror.Error)

//Handler is main handler for rpc over http
type Handler struct {
	log     *logrus.Entry
	methods map[string]HandlerMethod
}

//NewHandler init handler
func NewHandler(log *logrus.Entry) *Handler {
	return &Handler{
		log:     log.WithField("component", "httpServer"),
		methods: make(map[string]HandlerMethod),
	}
}

//AddMethod add method for name, we use name for identify method in rpc
func (h *Handler) AddMethod(name string, methodFunc HandlerMethod) error {
	if _, ok := h.methods[name]; ok {
		return errors.New("method with name already exists")
	}
	h.methods[name] = methodFunc
	h.log.Infof("HTTP method added: %s", name)
	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := h.log
	defer func() {
		if rec := recover(); rec != nil {
			resp := prepareErrRespBytes(
				serror.New(serror.ErrCodeBadRequest, fmt.Sprintf("Unexpected internal error, %+v", rec)),
			)
			writeResp(log, w, resp)
		}
	}()
	if r.Method != http.MethodPost {
		resp := prepareErrRespBytes(serror.New(serror.ErrCodeBadRequest, "Expected http method POST"))
		writeResp(log, w, resp)
		return
	}

	reqBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		resp := prepareErrRespBytes(serror.New(serror.ErrCodeBadRequest, "Can't read request body"))
		writeResp(log, w, resp)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.WithError(err).Error("Can't close body after reading")
	}
	log.WithField("data", string(reqBytes)).Trace("Request")

	var req *protocol.Request
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		resp := prepareErrRespBytes(serror.New(serror.ErrCodeBadRequest, "Body is not valid json"))
		writeResp(log, w, resp)
		return
	}
	log = log.WithField("method", req.Method)
	log.WithField("data", string(req.Data)).Debug("Request data")
	if _, ok := h.methods[req.Method]; !ok {
		resp := prepareErrRespBytes(
			serror.New(serror.ErrCodeNotFound, fmt.Sprintf("Unknown method: %s", req.Method)),
		)
		writeResp(log, w, resp)
		return
	}

	respData, serr := h.methods[req.Method](log, req.ID, req.Data)
	if serr != nil {
		resp := prepareErrRespBytes(serr)
		writeResp(log, w, resp)
		return
	}

	resp := protocol.Response{
		Type:   protocol.ResponseTypeOK,
		Result: respData,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		resp := prepareErrRespBytes(
			serror.New(serror.ErrCodeInternalError, "Can't marshal response"),
		)
		writeResp(log, w, resp)
		return
	}

	writeResp(log, w, respBytes)
}

func writeResp(log *logrus.Entry, w http.ResponseWriter, data []byte) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	log.WithField("data", string(data)).Trace("Response")
	if _, err := w.Write(data); err != nil {
		log.WithError(err).Error("Can't write response")
		return
	}
}

func prepareErrRespBytes(serr *serror.Error) []byte {
	bytes, err := json.Marshal(serr)
	if err != nil {
		return serror.GetMarshalErrorBytes()
	}
	res := protocol.Response{
		Type:   protocol.ResponseTypeErr,
		Result: bytes,
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		return protocol.GetMarshalResponseErrorBytes()
	}
	return resBytes
}
