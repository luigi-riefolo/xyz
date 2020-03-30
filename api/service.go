package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/luigi-riefolo/xyz/pb"
)

const (
	bearerKey                        = "bearer"
	firebaseServiceAccountKeyFileEnv = "FIREBASE_SERVICE_ACCOUNT_KEY_FILE"
	gwPort                           = 8080
	tlsPrivateKey                    = "certs/private.key"
	projectsID                       = "projects"
	svcPort                          = 9090
	tlsCertificate                   = "certs/certificate.crt"
)

var (
	grpcServerOpenEndpoint = flag.String(
		"xyz-open-grpc-server",
		"localhost:9090",
		"XYZ gRPC server endpoint")

	grpcServerEndpoint = flag.String(
		"xyz-grpc-server",
		"localhost:9090",
		"XYZ gRPC server endpoint")
)

// Service represents the XYZ API service.
type Service struct {
	gwmux  *runtime.ServeMux
	mux    *http.ServeMux
	server *grpc.Server

	// cert is a self signed certificate
	cert tls.Certificate
	// certPool contains the self signed certificate
	certPool *x509.CertPool

	authClient *auth.Client
	dbClient   *firestore.Client

	projects *firestore.CollectionRef
}

// AddTLS sets up TLS for the protected API endpoints.
func (s *Service) AddTLS(ctx context.Context) error {

	cert, err := ioutil.ReadFile(tlsCertificate)
	if err != nil {
		return errors.Wrap(err, "failed to read TLS certificate file")
	}

	key, err := ioutil.ReadFile(tlsPrivateKey)
	if err != nil {
		return errors.Wrap(err, "failed to read TLS private key file")
	}

	s.cert, err = tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return errors.Wrap(err, "failed to parse TLS key pair")
	}

	s.cert.Leaf, err = x509.ParseCertificate(s.cert.Certificate[0])
	if err != nil {
		return errors.Wrap(err, "failed to parse TLS certificate")
	}

	s.certPool = x509.NewCertPool()
	s.certPool.AddCert(s.cert.Leaf)

	return nil
}

// NewXYZService returns a ready-to-use XYZ service.
func NewXYZService(ctx context.Context) (pb.XYZServer, error) {

	firebaseServiceAccountKeyFile, ok := os.LookupEnv(firebaseServiceAccountKeyFileEnv)
	if !ok {
		return nil,
			fmt.Errorf("please provide a Firebase service account key file using the %s environemnt variable",
				firebaseServiceAccountKeyFileEnv)
	}

	app, err := firebase.NewApp(
		ctx,
		nil,
		option.WithCredentialsFile(firebaseServiceAccountKeyFile))
	if err != nil {
		return nil, errors.Wrap(err, "could not initialise app")
	}

	svc := &Service{}

	// register Firebase clients
	svc.authClient, err = app.Auth(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not create auth client")
	}

	svc.dbClient, err = app.Firestore(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not create db client")
	}

	svc.projects = svc.dbClient.Collection(projectsID)

	svc.mux = http.NewServeMux()

	svc.AddTLS(ctx)

	return svc, nil
}

// Start the server.
func (s *Service) Start(ctx context.Context) error {

	grpcGwAddress := net.JoinHostPort("0.0.0.0", strconv.Itoa(gwPort))
	svcAddress := net.JoinHostPort("0.0.0.0", strconv.Itoa(svcPort))

	lis, err := net.Listen("tcp", svcAddress)
	if err != nil {
		return errors.Wrap(err, "could not set up service listener")
	}

	s.server = grpc.NewServer(
		grpc.Creds(credentials.NewServerTLSFromCert(&s.cert)),
	)

	// register gRPC servers
	pb.RegisterOpenXYZServer(s.server, s)
	pb.RegisterXYZServer(s.server, s)

	go func() {
		log.Fatal(s.server.Serve(lis))
	}()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(s.certPool, "")),
	}

	s.gwmux = runtime.NewServeMux(
		// This is necessary to get error details properly
		// marshalled in unary requests.
		runtime.WithProtoErrorHandler(runtime.DefaultHTTPProtoErrorHandler),
	)

	// register gRPC server open endpoint
	err = pb.RegisterOpenXYZHandlerFromEndpoint(ctx, s.gwmux, *grpcServerOpenEndpoint, opts)
	if err != nil {
		return errors.Wrap(err, "could not register API handlers")
	}

	chain := []grpc.UnaryClientInterceptor{
		BearerAuthUnaryClientInterceptor(s.authClient),
	}

	// add authentication middleware
	opts = append(opts,
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(chain...)))

	// register gRPC server protected endpoint
	err = pb.RegisterXYZHandlerFromEndpoint(ctx, s.gwmux, *grpcServerEndpoint, opts)
	if err != nil {
		return errors.Wrap(err, "could not register API handlers")
	}

	s.mux.Handle("/", s.gwmux)

	gwServer := http.Server{
		Addr: fmt.Sprintf("localhost:%d", gwPort),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{s.cert},
		},
		Handler: s.mux,
	}

	log.Printf("XYZ HTTP gateway listening on %v", grpcGwAddress)

	// start HTTP server (and proxy calls to gRPC server endpoint)
	return gwServer.ListenAndServeTLS("", "")
}
