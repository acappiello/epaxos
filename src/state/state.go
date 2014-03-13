// Package state stores and updates the current sate of the replica.
package state

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"time"

	"commands"
	"datatypes"
	"listener"
	"message"
	"replicainfo"
)

// BATCHSIZE is the maximum number of Commands to send in a single Message.
const BATCHSIZE int = 100

// CHANSIZE is the size that channels are created with.
// Currently, will deadlock if this is too small.
const CHANSIZE int = 100000

// State holds all vital information about the replica.
type State struct {
	// Self is used to identify this replica in outgoing Messages.
	Self replicainfo.ReplicaInfo
	// Information about all other known replicas.
	Peers []replicainfo.ReplicaInfo
	// PeerMap is used to get the index of a particular replica in the slices.
	PeerMap map[int]int
	// Connections is a cache of sockets to other replicas.
	Connections []net.Conn
	// Readers is a cache of buffered readers to other replicas.
	Readers []*bufio.Reader
	// Writers is a cache of buffered writers to other replicas.
	Writers []*bufio.Writer

	Data *commands.Data

	// nPeers is the number of other replicas.
	nPeers int
	quorum int
	// instance is the current instance for this replica.
	instance uint32

	serverSocket net.Listener
	listener     *listener.Listener

	// Channels for unprocessed tasks.
	clientCommandsIn    chan *commands.Command
	preacceptCommandsIn chan *commands.Command
	acceptCommandsIn    chan *commands.Command
	commitCommandsIn    chan *commands.Command
	okCommandsIn        chan *commands.Command
	adminCommandsIn     chan *commands.Command

	// Channels for tasks that have not yet been sent.
	clientCommandsOut    chan *commands.Command
	preacceptCommandsOut chan *commands.Command
	acceptCommandsOut    chan *commands.Command
	commitCommandsOut    chan *commands.Command
	okCommandsOut        chan *commands.Command
	adminCommandsOut     chan *commands.Command

	executeCommands         chan *commands.Command
	priorityExecuteCommands chan *commands.Command
}

// Initialize sets the startup state of the repica.
func Initialize(port int, nreplica int) (s *State, err error) {
	s = new(State)
	err = nil
	// TODO: Handle error.
	host, _ := os.Hostname()
	s.Self.Hostname = []byte(host)
	s.Self.Port = port
	s.Self.Id = nreplica - 1
	s.Peers = make([]replicainfo.ReplicaInfo, nreplica-1)
	s.PeerMap = make(map[int]int)
	s.Connections = make([]net.Conn, nreplica-1)
	s.Readers = make([]*bufio.Reader, nreplica-1)
	s.Writers = make([]*bufio.Writer, nreplica-1)

	s.nPeers = nreplica - 1
	s.quorum = s.nPeers
	s.instance = 1

	s.serverSocket, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return
	}
	s.listener = listener.NewListener(s.serverSocket)

	s.clientCommandsIn = make(chan *commands.Command, CHANSIZE)
	s.preacceptCommandsIn = make(chan *commands.Command, CHANSIZE)
	s.acceptCommandsIn = make(chan *commands.Command, CHANSIZE)
	s.commitCommandsIn = make(chan *commands.Command, CHANSIZE)
	s.okCommandsIn = make(chan *commands.Command, CHANSIZE)
	s.adminCommandsIn = make(chan *commands.Command, CHANSIZE)

	s.clientCommandsOut = make(chan *commands.Command, CHANSIZE)
	s.preacceptCommandsOut = make(chan *commands.Command, CHANSIZE)
	s.acceptCommandsOut = make(chan *commands.Command, CHANSIZE)
	s.commitCommandsOut = make(chan *commands.Command, CHANSIZE)
	s.okCommandsOut = make(chan *commands.Command, CHANSIZE)
	s.adminCommandsOut = make(chan *commands.Command, CHANSIZE)

	s.executeCommands = make(chan *commands.Command, CHANSIZE)
	s.priorityExecuteCommands = make(chan *commands.Command, CHANSIZE)

	gob.Register(datatypes.Slot{})

	return
}

func (s *State) Status() {
	for {
		fmt.Println("-------------------------------")
		fmt.Println("clientCommandsIn: ", len(s.clientCommandsIn))
		fmt.Println("preacceptCommandsIn: ", len(s.preacceptCommandsIn))
		fmt.Println("acceptCommandsIn: ", len(s.acceptCommandsIn))
		fmt.Println("commitCommandsIn: ", len(s.commitCommandsIn))
		fmt.Println("okCommandsIn: ", len(s.okCommandsIn))
		fmt.Println("adminCommandsIn: ", len(s.adminCommandsIn))
		fmt.Println("clientCommandsOut: ", len(s.clientCommandsOut))
		fmt.Println("preacceptCommandsOut: ", len(s.preacceptCommandsOut))
		fmt.Println("acceptCommandsOut: ", len(s.acceptCommandsOut))
		fmt.Println("commitCommandsOut: ", len(s.commitCommandsOut))
		fmt.Println("okCommandsOut: ", len(s.okCommandsOut))
		fmt.Println("adminCommandsOut: ", len(s.adminCommandsOut))
		time.Sleep(time.Second)
	}
}

// Run is called once the replica is ready to receive input.
func (s *State) Run() {
	go s.ProcessIncoming()
	go s.ProcessOutgoing()
	go s.ProcessExecute()
	go s.listener.Listen()
	//go s.Status()
	for {
		m := s.listener.Get()
		//fmt.Println(m)
		s.AddCommands(m)
	}
}

// Registers a new peer, caching the connection and starting a goroutine
// to wait for incoming messages.
func (s *State) registerConnection(conn net.Conn, i int) {
	s.Connections[i] = conn
	s.Readers[i] = bufio.NewReader(conn)
	s.Writers[i] = bufio.NewWriter(conn)
	go s.listener.HandleConnection(conn)
}

// WaitForPeers is called by a single replica in the group that initially acts
// as a coordinator.
func (s *State) WaitForPeers() {
	m := &message.Message{}
	nextPeer := 0
	for i := 0; i < s.nPeers; i++ {
		conn, err := s.serverSocket.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Bad connection.")
			continue
		}
		s.registerConnection(conn, nextPeer)
		m.Unmarshal(s.Readers[nextPeer])
		fmt.Printf("%s:%d\n", string(m.Rep.Hostname), m.Rep.Port)
		m.Rep.Id = nextPeer
		s.Peers[nextPeer] = m.Rep
		s.PeerMap[nextPeer] = nextPeer
		nextPeer++
	}
}

// GetPeers is called by all replicas except the initial coordinator.
// The coordinator will send information about all other replicas.
func (s *State) GetPeers(wire net.Conn) {
	buf := bufio.NewReader(wire)
	m := &message.Message{}
	m.Unmarshal(buf)
	s.Self.Id = m.Rep.Id
	nextPeer := 0
	for i := 0; i < s.nPeers; i++ {
		m.Unmarshal(buf)
		fmt.Printf("%s:%d\n", string(m.Rep.Hostname), m.Rep.Port)
		if i > 0 {
			conn, err := net.Dial("tcp",
				fmt.Sprintf("%s:%d", string(m.Rep.Hostname), m.Rep.Port))
			if err != nil {
				fmt.Fprintln(os.Stderr, "Bad connection:", err)
			}
			s.registerConnection(conn, nextPeer)
		} else {
			// Reuse the existing connection.
			s.Connections[i] = wire
			s.Readers[i] = buf
			s.Writers[i] = bufio.NewWriter(wire)
		}
		s.Peers[i] = m.Rep
		s.PeerMap[m.Rep.Id] = i
		nextPeer++
	}
	go s.listener.HandleConnection(wire)
}

// SendHosts sends each peer information about all other replicas
// (including this one).
func (s *State) SendHosts() {
	for i := 0; i < s.nPeers; i++ {
		buf := s.Writers[i]
		m := message.AddReplica(s.Peers[i])
		m.Send(buf)
		m = message.AddReplica(s.Self)
		m.Send(buf)
		for j := 0; j < s.nPeers; j++ {
			// Don't send information about this peer to itself.
			if i != j {
				m = message.AddReplica(s.Peers[j])
				m.Send(buf)
			}
		}
	}
}

// AddCommands sorts incoming tasks into the proper channels.
func (s *State) AddCommands(m message.Message) {
	for _, t := range m.Commands {
		// TODO: Shouldn't need to allocate new memory.
		cmd := new(commands.Command)
		*cmd = t
		switch m.T {
		case message.REQUEST:
			s.clientCommandsIn <- cmd
		case message.CONNECT:
			s.adminCommandsIn <- cmd
		case message.HOSTLIST:
			s.adminCommandsIn <- cmd
		case message.ADDHOST:
			s.adminCommandsIn <- cmd
		case message.PREACCEPT:
			s.preacceptCommandsIn <- cmd
		case message.OK:
			s.okCommandsIn <- cmd
		case message.ACCEPT:
			s.acceptCommandsIn <- cmd
		case message.COMMIT:
			s.commitCommandsIn <- cmd
		}
	}
}

func (s *State) commit(t *commands.Command) {
	// TODO: Case on read vs execute.
	s.executeCommands <- t
	s.commitCommandsOut <- t
	s.sendToClient(t)
}

// ProcessIncoming takes an appropriate action for unhandled tasks.
// Channels are prioritized.
func (s *State) ProcessIncoming() {
	for {
		select {
		//case t := <-s.adminCommandsIn:
		case t := <-s.commitCommandsIn:
			//fmt.Println("GOT COMMIT: ", t)
			s.Data.HandleCommit(t)
			s.executeCommands <- t
		case t := <-s.okCommandsIn:
			if t.Accepted {
				//fmt.Println("GOT ACCEPT OK: ", t)
				noks := s.Data.HandleAcceptOk(t)
				if noks >= s.quorum {
					t.Committed = true
					s.commit(t)
				}
			} else {
				//fmt.Println("GOT PREACCEPT OK: ", t)
				cmd := s.Data.HandlePreacceptOk(t)
				//fmt.Println(noks, s.quorum)
				if cmd.NOks >= s.quorum {
					if cmd.Slow {
						cmd.Accepted = true
						cmd.NOks = 0
						s.acceptCommandsOut <- cmd
					} else {
						t.Accepted = true
						t.Committed = true
						s.commit(cmd)
					}
				}
			}
		case t := <-s.acceptCommandsIn:
			//fmt.Println("GOT ACCEPT: ", t)
			s.Data.HandleAccept(t)
			s.okCommandsOut <- t
		case t := <-s.preacceptCommandsIn:
			//fmt.Println("GOT PREACCEPT: ", t)
			s.Data.HandlePreaccept(t)
			s.okCommandsOut <- t
		case t := <-s.clientCommandsIn:
			//fmt.Println("GOT CLIENT REQ: ", t)
			t.S.ReplicaId = s.Self.Id
			t.S.Inst = s.instance
			s.instance++
			s.Data.AddDepsAndSeq(t)
			s.preacceptCommandsOut <- t
		}
	}
}

func (s *State) sendToClient(c *commands.Command) {
	w, exists := s.listener.ClientMap[c.ClientId]
	if !exists {
		return
	}
	m := &message.Message{
		T:        message.COMMIT,
		Rep:      s.Self,
		Commands: make([]commands.Command, 1),
	}
	m.Commands[0] = *c
	m.Send(w)
}

// batch reads from the given channel either until BATCHSIZE reads or no
// tasks remain. The first task is already provided.
func batch(ch chan *commands.Command, t1 *commands.Command) []commands.Command {
	tasks := make([]commands.Command, BATCHSIZE)
	tasks[0] = *t1
	for i := 1; i < BATCHSIZE; i++ {
		select {
		case t := <-ch:
			tasks[i] = *t
		default:
			return tasks[:i]
		}
	}
	return tasks
}

// sendReply sends a Message with a single Command to the replica indicated by
// t.HostId.
func (s *State) sendReply(t *commands.Command) {
	writer := s.Writers[s.PeerMap[t.S.ReplicaId]]
	m := &message.Message{
		T:        message.OK,
		Rep:      s.Self,
		Commands: make([]commands.Command, 1),
	}
	m.Commands[0] = *t
	m.Send(writer)
}

// sendToAll sends a Message with a group of Commands to all peers.
func (s *State) sendToAll(tsk []commands.Command, t message.MsgType) {
	m := &message.Message{
		T:        t,
		Rep:      s.Self,
		Commands: tsk,
	}
	for _, w := range s.Writers {
		m.Send(w)
	}
}

// ProcessOutgoing sends Messages for completed Commands.
func (s *State) ProcessOutgoing() {
	for {
		var tsk []commands.Command
		select {
		//case t := <-s.adminCommandsOut:
		case t := <-s.okCommandsOut:
			//fmt.Println("SEND OK: ", t)
			s.sendReply(t)
		case t := <-s.commitCommandsOut:
			tsk = batch(s.commitCommandsOut, t)
			//fmt.Println("SEND COMMIT", tsk)
			s.sendToAll(tsk, message.COMMIT)
		case t := <-s.acceptCommandsOut:
			tsk = batch(s.acceptCommandsOut, t)
			//fmt.Println("SEND ACCEPT", tsk)
			s.sendToAll(tsk, message.ACCEPT)
		case t := <-s.preacceptCommandsOut:
			tsk = batch(s.preacceptCommandsOut, t)
			//fmt.Println("SEND PREACCEPT", tsk)
			s.sendToAll(tsk, message.PREACCEPT)
			//case t:= <-s.clientCommandsOut:
		}
	}
}

func (s *State) ProcessExecute() {
	var cmd *commands.Command
	go s.Data.ExecuteComponents()
	for {
		select {
		case t := <-s.priorityExecuteCommands:
			cmd = t
		case t := <-s.executeCommands:
			cmd = t
		}
		g := s.Data.BuildGraph(cmd)
		g.SCC()
	}
}
