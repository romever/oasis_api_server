package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	//"github.com/oasisprotocol/oasis-core/go/common/crypto/address"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"
	"net/http"

	"google.golang.org/grpc"

	lgr "github.com/SimplyVC/oasis_api_server/src/logger"
	"github.com/SimplyVC/oasis_api_server/src/responses"
	"github.com/SimplyVC/oasis_api_server/src/rpc"
	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	common_signature "github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	consensus "github.com/oasisprotocol/oasis-core/go/consensus/api"
	beacon "github.com/oasisprotocol/oasis-core/go/beacon/api"
	mint_api "github.com/oasisprotocol/oasis-core/go/consensus/tendermint/api"
	"github.com/oasisprotocol/oasis-core/go/consensus/tendermint/crypto"
)

// loadConsensusClient loads consensus client and returns it
func loadConsensusClient(socket string) (*grpc.ClientConn,
	consensus.ClientBackend) {

	// Attempt to load connection with consensus client
	connection, consensusClient, err := rpc.ConsensusClient(socket)
	if err != nil {
		lgr.Error.Println("Failed to establish connection to consensus"+
			" client : ", err)
		return nil, nil
	}
	return connection, consensusClient
}

// loadBeaconClient loads beacon client and returns it
func loadBeaconClient(socket string) (*grpc.ClientConn,
	beacon.Backend) {

	// Attempt to load connection with beacon client
	connection, beaconClient, err := rpc.BeaconClient(socket)
	if err != nil {
		lgr.Error.Println("Failed to establish connection to beacon"+
			" client : ", err)
		return nil, nil
	}
	return connection, beaconClient
}

// GetConsensusStateToGenesis returns genesis state
// at specified block height for Consensus.
func GetConsensusStateToGenesis(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")

	// Retrieving name of node from query request
	nodeName := r.URL.Query().Get("name")
	confirmation, socket := checkNodeName(nodeName)
	if confirmation == false {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Node name requested doesn't exist"})
		return
	}

	// Retrieve height from query
	recvHeight := r.URL.Query().Get("height")
	height := checkHeight(recvHeight)
	if height == -1 {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Unexepcted value found, height needs to be " +
				"string of int!"})
		return
	}

	// Attempt to load connection with consensus client
	connection, co := loadConsensusClient(socket)

	// Close connection once code underneath executes
	defer connection.Close()

	// If null object was retrieved send response
	if co == nil {

		// Stop code here faild to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to establish connection using socket: " +
				socket})
		return
	}

	// Retrieving genesis state of consensus object at specified height
	consensusGenesis, err := co.StateToGenesis(context.Background(), height)
	if err != nil {
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to get Genesis file of Block!"})

		lgr.Error.Println("Request at /api/consensus/genesis failed "+
			"to retrieve genesis file : ", err)
		return
	}

	// Responding with consensus genesis state object, retrieved above.
	lgr.Info.Println("Request at /api/consensus/genesis responding with" +
		" genesis file!")
	json.NewEncoder(w).Encode(responses.ConsensusGenesisResponse{
		GenJSON: consensusGenesis})
}

// GetEpoch returns current epoch of given block height
func GetEpoch(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")

	// Retrieving name of node from query request
	nodeName := r.URL.Query().Get("name")
	confirmation, socket := checkNodeName(nodeName)
	if confirmation == false {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Node name requested doesn't exist"})
		return
	}

	// Retrieve height from query
	recvHeight := r.URL.Query().Get("height")
	height := checkHeight(recvHeight)
	if height == -1 {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Unexpected value found, height needs to be " +
				"a string representing an int!"})
		return
	}

	// Attempt to load connection with beacon client
	connection, be := loadBeaconClient(socket)

	// Close connection once code underneath executes
	defer connection.Close()

	// If null object was retrieved send response
	if be == nil {
		// Stop code here faild to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to establish connection using socket: " +
				socket})
		return
	}

	// Return epcoh of specific height
	epoch, err := be.GetEpoch(context.Background(), height)
	if err != nil {
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to retrieve Epoch of Block!"})

		lgr.Error.Println("Request at /api/consensus/epoch failed to"+
			" retrieve Epoch : ", err)
		return
	}

	// Respond with retrieved epoch above
	lgr.Info.Println("Request at /api/consensus/epoch responding" +
		" with an Epoch!")
	json.NewEncoder(w).Encode(responses.EpochResponse{Ep: epoch})
}

// PingNode returns consensus block at specific height
// thus signifying that it was pinged.
func PingNode(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")
	lgr.Info.Println("Received request for /api/pingnode")

	// Retrieving name of node from query request
	nodeName := r.URL.Query().Get("name")
	confirmation, socket := checkNodeName(nodeName)
	if confirmation == false {
		lgr.Info.Println("Node name requested doesn't exist")
		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Node name requested doesn't exist"})
		return
	}

	// Setting height to latest
	height := consensus.HeightLatest

	// Attempt to load connection with consensus client
	connection, co := loadConsensusClient(socket)

	// Close connection once code underneath executes
	defer connection.Close()

	// If null object was retrieved send response
	if co == nil {

		// Stop code here faild to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to establish connection using socket: " +
				socket})
		return
	}

	// Making sure that the error being retrieved is nill, meaning that API
	// is pingable
	_, err := co.GetBlock(context.Background(), height)
	if err != nil {
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to ping node by retrieving highest " +
				"block height!"})

		lgr.Error.Println("Request at /api/pingnode failed to ping"+
			" node : ", err)
		return
	}

	// Responding with Pong response
	lgr.Info.Println("Request at /api/pingnode responding with Pong!")
	json.NewEncoder(w).Encode(responses.SuccessResponsed)
}

// GetBlock returns consensus block at specific height.
func GetBlock(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")

	// Retrieving name of node from query request
	nodeName := r.URL.Query().Get("name")
	confirmation, socket := checkNodeName(nodeName)
	if confirmation == false {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Node name requested doesn't exist"})
		return
	}

	// Retrieve height from query
	recvHeight := r.URL.Query().Get("height")
	height := checkHeight(recvHeight)
	if height == -1 {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Unexepcted value found, height needs to be " +
				"string of int!"})
		return
	}

	// Attempt to load connection with consensus client
	connection, co := loadConsensusClient(socket)

	// Close connection once code underneath executes
	defer connection.Close()

	// If null object was retrieved send response
	if co == nil {

		// Stop code here faild to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to establish connection using socket: " +
				socket})
		return
	}

	// Retrieve block at specific height from consensus client
	blk, err := co.GetBlock(context.Background(), height)
	if err != nil {
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to retrieve Block!"})

		lgr.Error.Println("Request at /api/consensus/block failed "+
			"to retrieve Block : ", err)
		return
	}

	// Responding with retrieved block
	lgr.Info.Println(
		"Request at /api/consensus/block responding with Block!")
	json.NewEncoder(w).Encode(responses.BlockResponse{Blk: blk})
}

// GetBlockHeader returns consensus block header at specific height
func GetBlockHeader(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")

	// Retrieving name of node from query request
	nodeName := r.URL.Query().Get("name")
	confirmation, socket := checkNodeName(nodeName)
	if confirmation == false {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Node name requested doesn't exist"})
		return
	}

	// Retrieving height from query
	recvHeight := r.URL.Query().Get("height")
	height := checkHeight(recvHeight)
	if height == -1 {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Unexepcted value found, height needs to be " +
				"string of int!"})
		return
	}

	// Attempt to load connection with consensus client
	connection, co := loadConsensusClient(socket)

	// Close connection once code underneath executes
	defer connection.Close()

	// If null object was retrieved send response
	if co == nil {

		// Stop code here faild to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to establish connection using socket: " +
				socket})
		return
	}

	// Retriving Block at specific height using Consensus client
	blk, err := co.GetBlock(context.Background(), height)
	if err != nil {
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to retrieve Block!"})

		lgr.Error.Println("Request at /api/consensus/blockheader "+
			"failed to retrieve Block : ", err)
		return
	}

	// Creating BlockMeta object
	var meta mint_api.BlockMeta
	if err := cbor.Unmarshal(blk.Meta, &meta); err != nil {
		lgr.Error.Println("Request at /api/consensus/blockheader "+
			"failed to Unmarshal Block Metadata : ", err)

		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to Unmarshal Block Metadata!"})
		return
	}

	// Responds with block header retrieved above
	lgr.Info.Println("Request at /api/consensus/blockheader responding " +
		"with Block Header!")
	json.NewEncoder(w).Encode(responses.BlockHeaderResponse{
		BlkHeader: meta.Header})
}

// GetBlockLastCommit returns consensus block last commit at specific height
func GetBlockLastCommit(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")

	// Retrieving name of node from query request
	nodeName := r.URL.Query().Get("name")
	confirmation, socket := checkNodeName(nodeName)
	if confirmation == false {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Node name requested doesn't exist"})
		return
	}

	// Retrieving height from query
	recvHeight := r.URL.Query().Get("height")
	height := checkHeight(recvHeight)
	if height == -1 {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Unexepcted value found, height needs to be " +
				"string of int!"})
		return
	}

	// Attempt to load connection with consensus client
	connection, co := loadConsensusClient(socket)

	// Close connection once code underneath executes
	defer connection.Close()

	// If null object was retrieved send response
	if co == nil {

		// Stop code here faild to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to establish connection using socket: " +
				socket})
		return
	}

	// Retrieve block at specific height from consensus client
	blk, err := co.GetBlock(context.Background(), height)
	if err != nil {
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to retrieve Block!"})

		lgr.Error.Println("Request at /api/consensus/blocklastcommit "+
			"failed to retrieve Block : ", err)
		return
	}

	// Creating BlockMeta object
	var meta mint_api.BlockMeta
	if err := cbor.Unmarshal(blk.Meta, &meta); err != nil {
		lgr.Error.Println("Request at /api/consensus/blocklastcommit "+
			"failed Unmarshal Block Metadata : ", err)
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to Unmarshal Block Metadata!"})
		return
	}
	// Responds with Block Last commit retrieved above
	lgr.Info.Println("Request at /api/consensus/blocklastcommit " +
		"responding with Block Last Commit!")
	json.NewEncoder(w).Encode(responses.BlockLastCommitResponse{
		BlkLastCommit: meta.LastCommit})
}

// PublicKeyToAddress accepts a Consensus Public Key and respond with
// crypto.address which is used to match consensus public keys with
// Tendermint Addresses
func PublicKeyToAddress(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")

	// Retrieving consensus public key from the query
	consensusKey := r.URL.Query().Get("consensus_public_key")
	if consensusKey == "" {
		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "No Consensus Key Provided"})
		return
	}
	consensusPublicKey := &signature.PublicKey{}

	err := consensusPublicKey.UnmarshalText([]byte(consensusKey))
	if err != nil {
		lgr.Error.Println("Request at /api/consensus/pubkeyaddress "+
			"failed to Unmarshal Consensus PublicKey : ", err)
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to Unmarshal Public Key!"})
		return
	}
	// Convert the consensusKey into a signature PublicKey
	tendermintKey := crypto.PublicKeyToTendermint(consensusPublicKey)
	cryptoAddress := tendermintKey.Address()
	// Responds with transactions retrieved above
	lgr.Info.Println("Request at /api/consensus/pubkeyaddress responding " +
		"with Tendermint Public Key Address!")
	json.NewEncoder(w).Encode(responses.TendermintAddress{
		TendermintAddress: &cryptoAddress})
}

// GetTransactions returns consensus block header at specific height
func GetTransactions(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")

	// Retrieving name of node from query request
	nodeName := r.URL.Query().Get("name")
	confirmation, socket := checkNodeName(nodeName)
	if confirmation == false {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Node name requested doesn't exist"})
		return
	}

	// Retrieving height from query
	recvHeight := r.URL.Query().Get("height")
	height := checkHeight(recvHeight)
	if height == -1 {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Unexepcted value found, height needs to be " +
				"string of int!"})
		return
	}

	// Attempt to load connection with consensus client
	connection, co := loadConsensusClient(socket)

	// Close connection once code underneath executes
	defer connection.Close()

	// If null object was retrieved send response
	if co == nil {

		// Stop code here faild to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to establish connection using socket: " +
				socket})
		return
	}

	// Use consensus client to retrieve transactions at specific block
	// height
	transactions, err := co.GetTransactions(context.Background(), height)
	if err != nil {
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to retrieve Transactions!"})

		lgr.Error.Println("Request at /api/consensus/transactions "+
			"failed to retrieve Transactions : ", err)
		return
	}

	// Responds with transactions retrieved above
	lgr.Info.Println("Request at /api/consensus/transactions responding" +
		"with all transactions in specified Block!")
	json.NewEncoder(w).Encode(responses.TransactionsResponse{
		Transactions: transactions})
}

// GetTransactionsWithResults returns consensus block header at specific height
func GetTransactionsWithResults(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")

	// Retrieving name of node from query request
	nodeName := r.URL.Query().Get("name")
	confirmation, socket := checkNodeName(nodeName)
	if confirmation == false {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Node name requested doesn't exist"})
		return
	}

	// Retrieving height from query
	recvHeight := r.URL.Query().Get("height")
	height := checkHeight(recvHeight)
	if height == -1 {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Unexepcted value found, height needs to be " +
				"string of int!"})
		return
	}

	// Attempt to load connection with consensus client
	connection, co := loadConsensusClient(socket)

	// Close connection once code underneath executes
	defer connection.Close()

	// If null object was retrieved send response
	if co == nil {

		// Stop code here faild to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to establish connection using socket: " +
				socket})
		return
	}

	// Use consensus client to retrieve transactions at specific block
	// height
	transactions, err := co.GetTransactionsWithResults(context.Background(), height)
	if err != nil {
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to retrieve Transactions!"})

		lgr.Error.Println("Request at /api/consensus/transactionswithresults "+
			"failed to retrieve Transactions : ", err)
		return
	}

	// Responds with transactions retrieved above
	lgr.Info.Println("Request at /api/consensus/transactionswithresults responding" +
		"with all transactions in specified Block!")
	json.NewEncoder(w).Encode(responses.TransactionsWithResultsResponse{
		TransactionsWithResults: transactions})
}

func PublicKeyToBech32Address(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")

	// Retrieving consensus public key from the query
	consensusKey := r.URL.Query().Get("consensus_public_key")
	if consensusKey == "" {
		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "No Consensus Key Provided"})
		return
	}
	var pubKey common_signature.PublicKey

	err := pubKey.UnmarshalText([]byte(consensusKey))
	if err != nil {
		lgr.Error.Println("Request at /api/consensus/pubkeybech32address "+
			"failed to Unmarshal Consensus PublicKey : ", err)
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to Unmarshal Public Key!"})
		return
	}

	//var AddressV0Context = address.NewContext("oasis-core/address: staking", 0)
	cryptoAddress := staking.NewAddress(pubKey)

	// Responds with transactions retrieved above
	lgr.Info.Println("Request at /api/consensus/pubkeybech32address responding " +
		"with Bech32 Address!")
	json.NewEncoder(w).Encode(responses.Bech32Address{
		Bech32Address: &cryptoAddress})
}

func Base64ToBech32Address(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")

	// Retrieving consensus public key from the query
	base64Address := r.URL.Query().Get("address")
	if base64Address == "" {
		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "No Consensus Key Provided"})
		return
	}

	b, err := base64.StdEncoding.DecodeString(base64Address)
	if err != nil {
		lgr.Error.Println("Request at /api/consensus/base64bech32address "+
			"failed to Unmarshal Consensus PublicKey : ", err)
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to Unmarshal Public Key!"})
		return
	}

	var cryptoAddress staking.Address
	if err := cryptoAddress.UnmarshalBinary(b); err != nil {
		lgr.Error.Println("Request at /api/consensus/base64bech32address "+
			"failed to Unmarshal Consensus PublicKey : ", err)
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to Unmarshal Public Key!"})
		return
	}

	// Responds with transactions retrieved above
	lgr.Info.Println("Request at /api/consensus/base64bech32address responding " +
		"with Bech32 Address!")
	json.NewEncoder(w).Encode(responses.Bech32Address{
		Bech32Address: &cryptoAddress})
}

// GetStatus
func GetStatus(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")

	// Retrieving name of node from query request
	nodeName := r.URL.Query().Get("name")
	confirmation, socket := checkNodeName(nodeName)
	if confirmation == false {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Node name requested doesn't exist"})
		return
	}

	// Attempt to load connection with consensus client
	connection, co := loadConsensusClient(socket)

	// Close connection once code underneath executes
	defer connection.Close()

	// If null object was retrieved send response
	if co == nil {

		// Stop code here faild to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to establish connection using socket: " +
				socket})
		return
	}

	st, err := co.GetStatus(context.Background())
	if err != nil {
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to retrieve Status!"})

		lgr.Error.Println("Request at /api/consensus/status failed "+
			"to retrieve Status : ", err)
		return
	}

	lgr.Info.Println(
		"Request at /api/consensus/status responding with Block!")
	json.NewEncoder(w).Encode(responses.StatusResponse{St: st})
}

// GetHeight
func GetHeight(w http.ResponseWriter, r *http.Request) {

	// Add header so that received knows they're receiving JSON
	w.Header().Add("Content-Type", "application/json")

	// Retrieving name of node from query request
	nodeName := r.URL.Query().Get("name")
	confirmation, socket := checkNodeName(nodeName)
	if confirmation == false {

		// Stop code here no need to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Node name requested doesn't exist"})
		return
	}

	// Attempt to load connection with consensus client
	connection, co := loadConsensusClient(socket)

	// Close connection once code underneath executes
	defer connection.Close()

	// If null object was retrieved send response
	if co == nil {

		// Stop code here faild to establish connection and reply
		json.NewEncoder(w).Encode(responses.ErrorResponse{
			Error: "Failed to establish connection using socket: " +
				socket})
		return
	}

	height := checkHeight("")

	lgr.Info.Println(
		"Request at /api/consensus/status responding with Block!")
	json.NewEncoder(w).Encode(responses.HeightResponse{Ht: height})
}
