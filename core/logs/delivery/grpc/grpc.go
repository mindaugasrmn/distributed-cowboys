package processorasync

import (
	"context"
	"log"
	"reflect"

	logUsecases "github.com/mindaugasrmn/distributed-cowboys/core/logs/usecases"
	pbLog "github.com/mindaugasrmn/distributed-cowboys/proto/logs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//server dependencies
type server struct {
	pbLog.UnsafeLogsServiceV1Server
	log logUsecases.LogUsecases
}

var registry = make(map[string]reflect.Type)

func NewLogsServiceV1(
	s *grpc.Server,
	log logUsecases.LogUsecases,
) {

	srv := &server{
		log: log,
	}

	pbLog.RegisterLogsServiceV1Server(s, srv)
	reflection.Register(s)
}

//CreatePmtFile validate normal trx
func (s *server) LogV1(ctx context.Context, req *pbLog.LogRequestV1) (*pbLog.LogResponseV1, error) {
	log.Println(req.Record)
	off, err := s.log.Append(string(req.Record.Value))
	if err != nil {
		return nil, err
	}
	return &pbLog.LogResponseV1{Offset: off}, err
}
