package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/makerdao/testchain-deployment/pkg/serror"
	"github.com/makerdao/testchain-deployment/pkg/service/protocol"
	natsio "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
)

//HandlerMethod - method will be subscribed to topics
type HandlerMethod func(log *logrus.Entry, ID string, requestBytes []byte) (response []byte, error *serror.Error)

type Server struct {
	log         *logrus.Entry
	cfg         *Config
	conn        *natsio.Conn
	syncMethods map[string]HandlerMethod
	//asyncMethods map[string]HandlerMethod
}

func New(log *logrus.Entry, cfg *Config) *Server {
	return &Server{
		log:         log,
		cfg:         cfg,
		syncMethods: make(map[string]HandlerMethod),
	}
}

//AddSyncMethod add sync method for name
func (s *Server) AddSyncMethod(name string, methodFunc HandlerMethod) error {
	if _, ok := s.syncMethods[name]; ok {
		return errors.New("method with name already exists")
	}
	s.syncMethods[name] = methodFunc
	return nil
}

func (s *Server) Run(log *logrus.Entry) error {
	var err error
	s.conn, err = natsio.Connect(
		s.cfg.Servers,
		natsio.MaxReconnects(s.cfg.MaxReconnect),
		natsio.ReconnectWait(time.Duration(s.cfg.ReconnectWaitSec)*time.Second),
	)
	if err != nil {
		return err
	}

	for name, methodFunc := range s.syncMethods {
		log.Infof("Register sync method %s", name)
		topic := fmt.Sprintf("%s.%s.*", s.cfg.TopicPrefix, name)
		_, err := s.conn.QueueSubscribe(topic, s.cfg.GroupName, s.getSyncMsgHandler(methodFunc))
		if err != nil {
			return err
		}
		if err := s.conn.Flush(); err != nil {
			return err
		}
		if err := s.conn.LastError(); err != nil {
			return err
		}
	}

	return nil
}

//func (s *Server) getAsyncMsgHandler(methodFunc HandlerMethod) natsio.MsgHandler {
//	return func(msg *natsio.Msg) {
//		log := s.log
//		defer func() {
//			if rec := recover(); rec != nil {
//				resp := prepareErrRespBytes(
//					serror.New(serror.ErrCodeBadRequest, fmt.Sprintf("Unexpected internal error, %+v", rec)),
//				)
//				if err := s.conn.Publish(msg.Reply, resp); err != nil {
//					log.WithError(err).
//						WithField("topic", msg.Reply).
//						Error("Can't publish response with err to chanel")
//				}
//			}
//		}()
//		topicParts := strings.Split(msg.Subject, ".")
//		reqID := topicParts[len(topicParts)-1]
//		log = log.WithField("topic", msg.Subject)
//		log.WithField("data", string(msg.Data)).Info("Request")
//		_, sErr := methodFunc(log, reqID, msg.Data)
//		if sErr != nil {
//			errBytes := prepareErrRespBytes(sErr)
//			log.WithField("data", string(errBytes)).Error("Response error")
//			if err := s.conn.Publish(msg.Reply, errBytes); err != nil {
//				log.WithError(err).
//					WithField("topic", msg.Reply).
//					Error("Can't publish response with err to chanel")
//			}
//		}
//	}
//}

func (s *Server) getSyncMsgHandler(methodFunc HandlerMethod) natsio.MsgHandler {
	return func(msg *natsio.Msg) {
		log := s.log
		defer func() {
			if rec := recover(); rec != nil {
				resp := prepareErrRespBytes(
					serror.New(serror.ErrCodeBadRequest, fmt.Sprintf("Unexpected internal error, %+v", rec)),
				)
				if err := s.conn.Publish(msg.Reply, resp); err != nil {
					log.WithError(err).
						WithField("topic", msg.Reply).
						Error("Can't publish response with err to chanel")
				}
			}
		}()
		topicParts := strings.Split(msg.Subject, ".")
		reqID := topicParts[len(topicParts)-1]
		log = log.WithField("topic", msg.Subject)
		log.WithField("data", string(msg.Data)).Info("Request")
		res, sErr := methodFunc(log, reqID, msg.Data)
		if sErr != nil {
			errBytes := prepareErrRespBytes(sErr)
			log.WithField("data", string(errBytes)).Error("Response error")
			if err := s.conn.Publish(msg.Reply, errBytes); err != nil {
				log.WithError(err).
					WithField("topic", msg.Reply).
					Error("Can't publish response with err to chanel")
			}
		}
		response := protocol.Response{
			Type:   protocol.ResponseTypeOK,
			Result: res,
		}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			errBytes := prepareErrRespBytes(sErr)
			log.WithField("data", string(errBytes)).Error("Response marshaling error")
			errTopic := fmt.Sprintf("%s.%s.%s", s.cfg.TopicPrefix, s.cfg.ErrorTopic, reqID)
			if err := s.conn.Publish(errTopic, errBytes); err != nil {
				log.WithError(err).
					WithField("topic", msg.Reply).
					Error("Can't publish response with err to chanel")
			}
		}
		log.WithField("data", string(responseBytes)).Info("Response")
		if err := s.conn.Publish(msg.Reply, responseBytes); err != nil {
			log.WithError(err).
				WithField("topic", msg.Reply).
				Error("Can't publish response with err to chanel")
		}
	}
}

func (s *Server) Shutdown(context.Context, *logrus.Entry) error {
	return nil
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
