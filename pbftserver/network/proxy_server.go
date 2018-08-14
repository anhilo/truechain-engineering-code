package network

import (
	"math/big"
	"net/http"
	"github.com/truechain/truechain-engineering-code/pbftserver/consensus"
	"github.com/truechain/truechain-engineering-code/core/types"
	"encoding/json"
	"fmt"
	"bytes"
)


type Server struct {
	url string
	node *Node
	ID	*big.Int
	help consensus.ConsensusHelp
	server *http.Server
}

func NewServer(nodeID string,id *big.Int,help consensus.ConsensusHelp,
	verify consensus.ConsensusVerify,addrs []*types.CommitteeNode) *Server {
	if len(addrs) <= 0{
		return nil
	}
	node := NewNode(nodeID,verify,addrs)
		
	server := &Server{url:node.NodeTable[nodeID], node:node,ID:id,help:help,}
	server.server = &http.Server{
		Addr:		server.url,
	}
	server.setRoute()
	return server
}
func (server *Server) Start() {
	go server.startHttpServer()
}
func (server *Server) startHttpServer() {
	fmt.Printf("Server will be started at %s...\n", server.url)
	if err := server.server.ListenAndServe(); err != nil {
		fmt.Println(err)
		return
	}
}
func (server *Server) Stop(){
	// do nothing
	if server.server != nil {
		server.server.Close()
	}
}
func (server *Server) setRoute() {
	mux := http.NewServeMux()
	mux.HandleFunc("/req", server.getReq)
	mux.HandleFunc("/preprepare", server.getPrePrepare)
	mux.HandleFunc("/prepare", server.getPrepare)
	mux.HandleFunc("/commit", server.getCommit)
	mux.HandleFunc("/reply", server.getReply)
	server.server.Handler = mux
}

func (server *Server) getReq(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.RequestMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getPrePrepare(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.PrePrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getPrepare(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.VoteMsg
	var tmp consensus.StorgePrepareMsg
	err := json.NewDecoder(request.Body).Decode(&tmp)
	if err != nil {
		fmt.Println(err)
		return
	}
	msg.Digest,msg.NodeID,msg.ViewID = tmp.Digest,tmp.NodeID,tmp.ViewID
	msg.SequenceID,msg.MsgType = tmp.SequenceID,tmp.MsgType
	msg.Pass = nil
	server.node.MsgEntrance <- &msg
}

func (server *Server) getCommit(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.VoteMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getReply(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.ReplyMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	// server.node.GetReply(&msg)
	server.handleResult(&msg)
}
func (server *Server) handleResult(msg *consensus.ReplyMsg) {
	var res uint = 0
	if msg.NodeID == "Executed" {
		res = 1
	}
	if msg.ViewID == server.node.CurrentState.ViewID {
		server.help.ReplyResult(server.node.CurrentState.MsgLogs.ReqMsg,res)
	} else {
		// wrong state
	}
}
func (server *Server) PutRequest(msg *consensus.RequestMsg) {
	server.node.MsgEntrance <- msg
	height := big.NewInt(msg.Height)
	ac := &consensus.ActionIn{
		AC:		consensus.ActionBroadcast,
		ID:		server.ID,
		Height:	height,
	}
	consensus.ActionChan <- ac
}

func send(url string, msg []byte) {
	buff := bytes.NewBuffer(msg)
	http.Post("http://" + url, "application/json", buff)
}