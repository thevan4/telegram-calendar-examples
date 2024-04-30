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

var (
	grpcPort        string
	httpPort        string
	grpcDialTimeout time.Duration
	httpDialTimeout time.Duration
)

func init() {
	log.SetFlags(0)
	grpcPort = os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
		logJSON("debug", "grpc port set default: 55555")
	}

	httpPort = os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
		logJSON("debug", "http port set default: 55555")
	}

	var err error
	grpcDialTimeoutRaw := os.Getenv("GRPC_DIAL_TIMEOUT")
	if grpcDialTimeoutRaw == "" {
		grpcDialTimeout = time.Second
		logJSON("debug", "grpc dial timeout set default: 1s")
	} else {
		grpcDialTimeout, err = time.ParseDuration(grpcDialTimeoutRaw)
		if err != nil {
			grpcDialTimeout = time.Second
			logJSON("warn", "grpc dial timeout set error: "+err.Error()+", set default: 1s")
		} else {
			logJSON("debug", "grpc dial timeout set: "+grpcDialTimeout.String())
		}
	}

	httpDialTimeoutRaw := os.Getenv("HTTP_DIAL_TIMEOUT")
	if httpDialTimeoutRaw == "" {
		httpDialTimeout = time.Second
		logJSON("debug", "http dial timeout set default: 1s")
	} else {
		httpDialTimeout, err = time.ParseDuration(httpDialTimeoutRaw)
		if err != nil {
			httpDialTimeout = time.Second
			logJSON("warn", "http dial timeout set error: "+err.Error()+", set default: 1s")
		} else {
			logJSON("debug", "http dial timeout set: "+httpDialTimeout.String())
		}
	}
}

func main() {
	ctx, cancelAtStart := context.WithCancel(context.Background())
	defer cancelAtStart()

	grpcServer := grpc.NewServer()
	serviceGrpc := service.NewTelegramCalendarGRPCService()
	pb.RegisterCalendarServiceServer(grpcServer, serviceGrpc)

	wg := new(sync.WaitGroup)
	wg.Add(2)

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

	logJSON("info", "http starting at port "+httpPort)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if errPbRegister := pb.RegisterCalendarServiceHandlerFromEndpoint(ctx, mux, ":"+grpcPort, opts); errPbRegister != nil {
		logJSON("fatal", "failed to register HTTP endpoint: "+errPbRegister.Error())
	}
	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: mux,
	}
	go func() {
		go waitForServerReady(httpPort, wg, httpDialTimeout)
		if errHTTPServe := httpServer.ListenAndServe(); errHTTPServe != nil {
			if !errors.Is(errHTTPServe, http.ErrServerClosed) {
				logJSON("fatal", "failed to serve: "+errHTTPServe.Error())
			}
		}
	}()

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
