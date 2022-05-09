package processorasync

import (
	"context"
	"fmt"
	"reflect"

	cowboyUsecases "github.com/mindaugasrmn/distributed-cowboys/core/cowboy/usecases"
	pbCowboy "github.com/mindaugasrmn/distributed-cowboys/proto/cowboy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//server dependencies
type server struct {
	pbCowboy.UnsafeCowboyV1Server
	cowboyUse cowboyUsecases.CowboyUsecases
}

var registry = make(map[string]reflect.Type)

func NewCowboyV1(
	s *grpc.Server,
	cowboyUse cowboyUsecases.CowboyUsecases,

) {

	srv := &server{}

	pbCowboy.RegisterCowboyV1Server(s, srv)
	reflection.Register(s)
}

//CreatePmtFile validate normal trx
func (s *server) ShootV1(ctx context.Context, req *pbCowboy.ShootRequestV1) (*pbCowboy.ShootResponseV1, error) {
	health := s.cowboyUse.GetHealth()
	if health == 0 {
		return &pbCowboy.ShootResponseV1{Msg: fmt.Sprintf("%s is dead already", s.cowboyUse.GetName())}, nil
	}
	s.cowboyUse.GotDamage(int(req.Damage))
	return &pbCowboy.ShootResponseV1{Msg: fmt.Sprintf("%s got %d damage, he is %s", s.cowboyUse.GetName(), s.cowboyUse.GetDamage(), s.cowboyUse.GetStatus())}, nil
}
