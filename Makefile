PROJECTNAME=$(shell basename "$(PWD)")
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
CMD_DIR=cmd
HOME_DIR=${HOME}
PROTO_DIR=proto

k8s-apply:
	@echo " > Initializing minikube"
	@kubectl apply -f k8s/rules.yaml
	@kubectl apply -f k8s/01-zookeeper.yaml
	@kubectl apply -f k8s/02-kafka.yaml
	@kubectl apply -f k8s/03-redis.yaml



proto-all:
	@echo "  >  Compiling gRPC proto files"
	@echo $(shell ls ${PROTO_DIR});
	@for f in $(shell ls ${PROTO_DIR}); \
	do \
	    echo "  >  Compiling $${f}"; \
		protoc --proto_path=proto/$${f}/ --proto_path=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=proto/$${f} --go-grpc_out=proto/$${f} $${f}.proto; \
	done
	@echo "  >  Done!"


get-logs:
	@kubectl exec logd-service -- cat log.txt

init:
	@eval $(minikube docker-env)
	@echo "  >  Reinializing"
	@echo "  >  Deleating pods"
	@kubectl delete pods -l app=cowboy
	@kubectl delete pods -l app=starter
	@kubectl delete pods -l app=logd
	@echo "  >  Rebiulding cointainers"
	@docker build -f docker/cowboy/Dockerfile -t cowboy .
	@docker build  -f docker/logd/Dockerfile -t logd .
	@docker build -f docker/starter/Dockerfile -t starter .
	@echo "  >  Done!"
	@echo "  >  Creating pods"
	@go run cmd/init/main.go
	@echo "  >  Done!"

