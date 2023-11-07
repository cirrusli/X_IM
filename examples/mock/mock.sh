#export GOPATH=C:/Users/X/go # Windows
export PATH=$PATH:$(go env GOPATH)/bin

go get -u github.com/golang/mock/gomock
go get -u github.com/golang/mock/mockgen

mockgen --source ./server.go -package X_IM -destination server_mock.go
mockgen --source ./storage.go -package X_IM -destination storage_mock.go
mockgen --source ./dispatcher.go -package X_IM -destination dispatcher_mock.go
