// Package state stores and updates the current sate of the replica.
package state

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"listener"
	"message"
	"replicainfo"
)

// BATCHSIZE is the maximum number of Tasks to send in a single Message.
const BATCHSIZE int = 100
// CHANSIZE is the size that channels are created with.
const CHANSIZE int = 100

// State holds all vital information about the replica.
type State struct {
	// Self is used to identify this replica in outgoing Messages.
	Self        replicainfo.ReplicaInfo
	// Information about all other known replicas.
	Peers       []replicainfo.ReplicaInfo
	// PeerMap is used to get the index of a particular replica in the slices.
	PeerMap     map[mapkey]int
	// Connections is a cache of sockets to other replicas.
	Connections []net.Conn
	// Readers is a cache of buffered readers to other replicas.
	Readers     []*bufio.Reader
	// Writers is a cache of buffered writers to other replicas.
	Writers     []*bufio.Writer

	// nPeers is the number of other replicas.
	nPeers      int
	// instance is the current instance for this replica.
	instance    uint32

	serverSocket net.Listener
	listener     *listener.Listener

	// Channels for unprocessed tasks.
	clientTasksIn    chan message.Task
	preacceptTasksIn chan message.Task
	acceptTasksIn    chan message.Task
	commitTasksIn    chan message.Task
	okTasksIn        chan message.Task
	adminTasksIn     chan message.Task

	// Channels for tasks that have not yet been sent.
	clientTasksOut    chan message.Task
	preacceptTasksOut chan message.Task
	acceptTasksOut    chan message.Task
	commitTasksOut    chan message.Task
	okTasksOut        chan message.Task
	adminTasksOut     chan message.Task
}

// mapkey is used as keys in PeerMap.
type mapkey struct {
	host string
	port int
}

// getkey will uniquely identify a replica in PeerMap.
func getkey(rep replicainfo.ReplicaInfo) mapkey {
	return mapkey{
		string(rep.Hostname),
		rep.Port,
	}
}

// Initialize sets the startup state of the repica.
func Initialize(port int, nreplica int) (s *State, err error) {
	s = new(State)
	err = nil
	// TODO: Handle error.
	host, _ := os.Hostname()
	s.Self.Hostname = []byte(host)
	s.Self.Port = port
	s.Peers = make([]replicainfo.ReplicaInfo, nreplica-1)
	s.PeerMap = make(map[mapkey]int)
	s.Connections = make([]net.Conn, nreplica-1)
	s.Readers = make([]*bufio.Reader, nreplica-1)
	s.Writers = make([]*bufio.Writer, nreplica-1)
	s.nPeers = nreplica - 1

	s.serverSocket, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return
	}
	s.listener = listener.NewListener(s.serverSocket)

	s.clientTasksIn = make(chan message.Task, CHANSIZE)
	s.preacceptTasksIn = make(chan message.Task, CHANSIZE)
	s.acceptTasksIn = make(chan message.Task, CHANSIZE)
	s.commitTasksIn = make(chan message.Task, CHANSIZE)
	s.okTasksIn = make(chan message.Task, CHANSIZE)
	s.adminTasksIn = make(chan message.Task, CHANSIZE)

	s.clientTasksOut = make(chan message.Task, CHANSIZE)
	s.preacceptTasksOut = make(chan message.Task, CHANSIZE)
	s.acceptTasksOut = make(chan message.Task, CHANSIZE)
	s.commitTasksOut = make(chan message.Task, CHANSIZE)
	s.okTasksOut = make(chan message.Task, CHANSIZE)
	s.adminTasksOut = make(chan message.Task, CHANSIZE)

	return
}

// Run is called once the replica is ready to receive input.
func (s *State) Run() {
	go s.ProcessIncoming()
	go s.ProcessOutgoing()
	go s.listener.Listen()
	for {
		m := s.listener.Get()
		//fmt.Println(m)
		s.AddTasks(m)
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
		s.Peers[nextPeer] = m.Rep
		s.PeerMap[getkey(m.Rep)] = nextPeer
		nextPeer++
	}
}

// GetPeers is called by all replicas except the initial coordinator.
// The coordinator will send information about all other replicas.
func (s *State) GetPeers(wire net.Conn) {
	buf := bufio.NewReader(wire)
	m := &message.Message{}
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
			fmt.Printf("i: %d\n", i)
			s.Connections[i] = wire
			s.Readers[i] = buf
			s.Writers[i] = bufio.NewWriter(wire)
		}
		s.Peers[i] = m.Rep
		s.PeerMap[getkey(m.Rep)] = i
		nextPeer++
	}
	go s.listener.HandleConnection(wire)
}

// SendHosts sends each peer information about all other replicas
// (including this one).
func (s *State) SendHosts() {
	for i := 0; i < s.nPeers; i++ {
		buf := s.Writers[i]
		m := message.AddHost(string(s.Self.Hostname), s.Self.Port)
		m.Send(buf)
		for j := 0; j < s.nPeers; j++ {
			// Don't send information about this peer to itself.
			if i != j {
				m = message.AddHost(string(s.Peers[j].Hostname),
					s.Peers[j].Port)
				m.Send(buf)
			}
		}
	}
}

// AddTasks sorts incoming tasks into the proper channels.
func (s *State) AddTasks(m message.Message) {
	for _, t := range m.Tasks {
		switch m.T {
		case message.REQUEST:
			s.clientTasksIn <- t
		case message.CONNECT:
			s.adminTasksIn <- t
		case message.HOSTLIST:
			s.adminTasksIn <- t
		case message.ADDHOST:
			s.adminTasksIn <- t
		case message.PREACCEPT:
			t.HostId = s.PeerMap[getkey(m.Rep)]
			s.preacceptTasksIn <- t
		case message.PREACCEPTOK:
			s.okTasksIn <- t
		case message.ACCEPT:
			t.HostId = s.PeerMap[getkey(m.Rep)]
			s.acceptTasksIn <- t
		case message.ACCEPTOK:
			s.okTasksIn <- t
		case message.COMMIT:
			s.commitTasksIn <- t
		}
	}
}

// ProcessIncoming takes an appropriate action for unhandled tasks.
// Channels are prioritized.
func (s *State) ProcessIncoming() {
	for {
		select {
		//case t := <-s.adminTasksIn:
		case t := <-s.okTasksIn:
			fmt.Println("GOT PREACCEPT OK: ", t)
			//case t := <-s.commitTasksIn:
			//case t := <-s.acceptTasksIn:
		case t := <-s.preacceptTasksIn:
			fmt.Println("GOT PREACCEPT: ", t)
			s.okTasksOut <- t
		case t := <-s.clientTasksIn:
			fmt.Println("GOT CLIENT REQ: ", t)
			s.preacceptTasksOut <- t
		}
	}
}

// batch reads from the given channel either until BATCHSIZE reads or no
// tasks remain. The first task is already provided.
func batch(ch chan message.Task, t1 message.Task) []message.Task {
	end := false
	tasks := make([]message.Task, BATCHSIZE)
	tasks[0] = t1
	i := 1
	for ; i < BATCHSIZE && !end; i++ {
		select {
		case t := <-ch:
			tasks[i] = t
		default:
			end = true
		}
	}
	return tasks[:i-1]
}

// sendReply sends a Message with a single Task to the replica indicated by
// t.HostId.
func (s *State) sendReply(t message.Task) {
	writer := s.Writers[t.HostId]
	m := &message.Message{
		T:     message.PREACCEPTOK,
		Rep:   s.Self,
		Tasks: make([]message.Task, 1),
	}
	m.Tasks[0] = t
	m.Send(writer)
}

// sendToAll sends a Message with a group of Tasks to all peers.
func (s *State) sendToAll(tsk []message.Task) {
	m := &message.Message{
		T:     message.PREACCEPT,
		Rep:   s.Self,
		Tasks: tsk,
	}
	for _, w := range s.Writers {
		m.Send(w)
	}
}

// ProcessOutgoing sends Messages for completed Tasks.
func (s *State) ProcessOutgoing() {
	for {
		var tsk []message.Task
		select {
		//case t := <-s.adminTasksOut:
		case t := <-s.okTasksOut:
			fmt.Println("SEND PREACCEPT OK: ", t)
			s.sendReply(t)
			//case t := <-s.commitTasksOut:
			//case t := <-s.acceptTasksOut:
		case t := <-s.preacceptTasksOut:
			tsk = batch(s.preacceptTasksOut, t)
			fmt.Println("SEND PREACCEPT", tsk)
			s.sendToAll(tsk)
			//case t:= <-s.clientTasksOut:
		}
	}
}
