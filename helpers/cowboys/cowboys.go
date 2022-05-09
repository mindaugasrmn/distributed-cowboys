package cowboys

import (
	"encoding/json"
	"fmt"
	"os"

	cowboyDomain "github.com/mindaugasrmn/distributed-cowboys/core/cowboy/domain"
)

func CowboysArray(path string) ([]*cowboyDomain.Cowboy, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	var cowboys []*cowboyDomain.Cowboy
	if err := json.NewDecoder(fd).Decode(&cowboys); err != nil {
		return nil, fmt.Errorf("Failed to decode file, got error: %s", err.Error())
	}

	return cowboys, nil
}
