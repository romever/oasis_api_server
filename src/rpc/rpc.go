package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	beacon "github.com/oasisprotocol/oasis-core/go/beacon/api"
	cmnGrpc "github.com/oasisprotocol/oasis-core/go/common/grpc"
	"github.com/oasisprotocol/oasis-core/go/common/identity"
	consensus "github.com/oasisprotocol/oasis-core/go/consensus/api"
	control "github.com/oasisprotocol/oasis-core/go/control/api"
	governance "github.com/oasisprotocol/oasis-core/go/governance/api"
	registry "github.com/oasisprotocol/oasis-core/go/registry/api"
	scheduler "github.com/oasisprotocol/oasis-core/go/scheduler/api"
	sentry "github.com/oasisprotocol/oasis-core/go/sentry/api"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"
)

// SentryClient - initiate new sentry client
func SentryClient(address string, tlsPath string) (*grpc.ClientConn,
	sentry.Backend, error) {

	conn, err := ConnectTLS(address, tlsPath)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("Failed to establish Sentry "+
			"Connection with node %s", address)
	}

	client := sentry.NewSentryClient(conn)
	return conn, client, nil
}

// SchedulerClient - initiate new scheduler client
func SchedulerClient(address string) (*grpc.ClientConn, scheduler.Backend,
	error) {

	conn, err := Connect(address)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to establish Scheduler "+
			"Client Connection with node %s", address)
	}

	client := scheduler.NewSchedulerClient(conn)
	return conn, client, nil
}

// NodeControllerClient - initiate new registry client
func NodeControllerClient(address string) (*grpc.ClientConn,
	control.NodeController, error) {

	conn, err := Connect(address)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to establish "+
			"NodeController Connection with node %s", address)
	}

	client := control.NewNodeControllerClient(conn)
	return conn, client, nil
}

// RegistryClient - initiate new registry client
func RegistryClient(address string) (*grpc.ClientConn,
	registry.Backend, error) {

	conn, err := Connect(address)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to establish Registry "+
			"Client Connection with node %s", address)
	}

	client := registry.NewRegistryClient(conn)
	return conn, client, nil
}

// GovernanceClient - initiate new governance client
func GovernanceClient(address string) (*grpc.ClientConn,
	governance.Backend, error) {

	conn, err := Connect(address)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to establish Governance "+
			"Client Connection with node %s", address)
	}

	client := governance.NewGovernanceClient(conn)
	return conn, client, nil
}

// ConsensusClient - initiate new consensus client
func ConsensusClient(address string) (*grpc.ClientConn,
	consensus.ClientBackend, error) {
	conn, err := Connect(address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to establish connection "+
			"with node %s", address)
	}

	client := consensus.NewConsensusClient(conn)
	return conn, client, nil
}

// BeaconClient - initiate new beacon client
func BeaconClient(address string) (*grpc.ClientConn,
	beacon.Backend, error) {
	conn, err := Connect(address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to establish connection "+
			"with node %s", address)
	}

	client := beacon.NewBeaconClient(conn)
	return conn, client, nil
}

// ConsensusLightClient - initiate new consensus light client
func ConsensusLightClient(address string) (*grpc.ClientConn,
	consensus.LightClientBackend, error) {
	conn, err := Connect(address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to establish connection "+
			"with node %s", address)
	}

	client := consensus.NewConsensusLightClient(conn)
	return conn, client, nil
}

// StakingClient - initiate new staking client
func StakingClient(address string) (*grpc.ClientConn, staking.Backend, error) {
	conn, err := Connect(address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to establish connection "+
			"with node %s", address)
	}

	client := staking.NewStakingClient(conn)
	return conn, client, nil
}

// ConnectTLS connects to server using TLS Certificate
func ConnectTLS(address string, tlsPath string) (*grpc.ClientConn, error) {

	// Open and read tls file containing connection information
	b, err := ioutil.ReadFile(tlsPath)
	if err != nil {
		return nil, err
	}

	// Add Credentials to a certificate pool
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(b) {
		return nil, fmt.Errorf("credentials: failed to append " +
			"certificates")
	}

	// Create new TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		RootCAs:    certPool,
		ServerName: identity.CommonName,
	})

	// Add Credentials to grpc options to be used for TLS Connection
	opts := grpc.WithTransportCredentials(creds)
	conn, err := cmnGrpc.Dial(
		address,
		opts,
	)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Connect - connect to grpc
// Add grpc.WithBlock() and grpc.WithTimeout()
// to have dial to constantly try and establish connection
func Connect(address string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	opts = append(opts, grpc.WithDefaultCallOptions(
		grpc.WaitForReady(false)))

	conn, err := cmnGrpc.Dial(
		address,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
