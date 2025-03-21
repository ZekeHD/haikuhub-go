package internal

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getGrpcClientConn() *grpc.ClientConn {
	envLoadErr := godotenv.Load()
	if envLoadErr != nil {
		log.Fatal("Error loading env file", envLoadErr.Error())
	}

	grpcURI := os.Getenv("GRPC_URI")
	conn, err := grpc.NewClient(
		grpcURI,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		fmt.Println("Error establishing RPC server connection:", err.Error())
	}

	return conn
}

var GrpcClientConn = getGrpcClientConn()

/*
Now that I have Haiku validation working properly:

- implement mocks for RPC service in haikus_test.go
- implement filtering
*/
