package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/thevan4/telegram-calendar-examples/standalone_service/internal/service"
	pb "github.com/thevan4/telegram-calendar-examples/standalone_service/pkg/telegram-calendar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, cancelAtStart := context.WithCancel(context.Background())
	defer cancelAtStart()

	// start servers
	wg := new(sync.WaitGroup)
	wg.Add(2)
	grpcServer := startGRPCServer(wg)
	httpServer := startHTTPServer(ctx, wg)

	wg.Wait()
	logJSON("info", "gRPC (port "+grpcPort+") and HTTP (port "+httpPort+") servers are up and running")

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	logJSON("info", "stopping servers")
	ctxShutdown, cancelAtShutdown := context.WithTimeout(ctx, 5*time.Second)
	defer cancelAtShutdown()

	grpcStopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcStopped)
	}()

	if err := httpServer.Shutdown(ctxShutdown); err != nil {
		logJSON("error", "http server shutdown failed: "+err.Error())
	} else {
		logJSON("info", "http server stopped")
	}

	select {
	case <-grpcStopped:
		logJSON("info", "grpc server stopped gracefully")
	case <-ctxShutdown.Done():
		grpcServer.Stop()
		logJSON("warn", "grpc server stopped")
	}
}

func startGRPCServer(wg *sync.WaitGroup) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(newGRPCAuthInterceptor()),
	)
	serviceGrpc := service.NewTelegramCalendarGRPCService()
	pb.RegisterCalendarServiceServer(grpcServer, serviceGrpc)

	logJSON("info", "grpc starting at port "+grpcPort)
	listerGrpc, errTCPListen := net.Listen("tcp", ":"+grpcPort)
	if errTCPListen != nil {
		logJSON("fatal", "failed to listen: "+errTCPListen.Error())
	}
	go func() {
		go waitForServerReady(grpcPort, wg, grpcDialTimeout)
		if errGrpcServe := grpcServer.Serve(listerGrpc); errGrpcServe != nil {
			logJSON("fatal", "failed to serve gRPC: "+errGrpcServe.Error())
		}
	}()

	return grpcServer
}

func startHTTPServer(ctx context.Context, wg *sync.WaitGroup) *http.Server {
	logJSON("info", "http starting at port "+httpPort)
	httpMux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if errPbRegister := pb.RegisterCalendarServiceHandlerFromEndpoint(ctx, httpMux, ":"+grpcPort, opts); errPbRegister != nil {
		logJSON("fatal", "failed to register HTTP endpoint: "+errPbRegister.Error())
	}
	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: authHTTPInterceptor(httpMux),
	}
	go func() {
		go waitForServerReady(httpPort, wg, httpDialTimeout)
		if errHTTPServe := httpServer.ListenAndServe(); errHTTPServe != nil {
			if !errors.Is(errHTTPServe, http.ErrServerClosed) {
				logJSON("fatal", "failed to serve: "+errHTTPServe.Error())
			}
		}
	}()

	return httpServer
}

func waitForServerReady(port string, wg *sync.WaitGroup, dialTime time.Duration) {
	defer wg.Done()
	conn, err := net.DialTimeout("tcp", ":"+port, dialTime)
	if err != nil {
		logJSON("fatal", "at waitForServerReady dial to port "+port+" error: "+err.Error())
	}

	if errClose := conn.Close(); errClose != nil {
		logJSON("info", "at waitForServerReady port "+port+" close error: "+errClose.Error())
	}
}

func logJSON(level, message string) {
	currentTime := time.Now().Format(time.RFC3339)
	jsonMessage := `{"time":"%s", "level":"%s", "message":"%s"}`
	switch level {
	case "fatal":
		log.Fatalf(jsonMessage, currentTime, level, message)
	default:
		log.Printf(jsonMessage, currentTime, level, message)
	}
}
