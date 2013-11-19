package state

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"message"
	"replicainfo"
)

const BATCHSIZE int = 10
const CHANSIZE int = 100

type State struct {
	Self        replicainfo.ReplicaInfo
	Peers       []replicainfo.ReplicaInfo
	PeerMap     map[mapkey]int
	Connections []net.Conn
	Readers     []*bufio.Reader
	Writers     []*bufio.Writer
	nextPeer    int
	nPeers      int
	instance    uint32

	clientTasksIn    chan message.Task
	preacceptTasksIn chan message.Task
	acceptTasksIn    chan message.Task
	commitTasksIn    chan message.Task
	okTasksIn        chan message.Task
	adminTasksIn     chan message.Task

	clientTasksOut    chan message.Task
	preacceptTasksOut chan message.Task
	acceptTasksOut    chan message.Task
	commitTasksOut    chan message.Task
	okTasksOut        chan message.Task
	adminTasksOut     chan message.Task
}

type mapkey struct {
	host string
	port int
}

func getkey(rep replicainfo.ReplicaInfo) mapkey {
	return mapkey{
		string(rep.Hostname),
		rep.Port,
	}
}

func Initialize(port int, nreplica int) *State {
	s := new(State)
	// TODO: Handle error.
	host, _ := os.Hostname()
	s.Self.Hostname = []byte(host)
	s.Self.Port = port
	s.Peers = make([]replicainfo.ReplicaInfo, nreplica-1)
	s.PeerMap = make(map[mapkey]int)
	s.Connections = make([]net.Conn, nreplica-1)
	s.Readers = make([]*bufio.Reader, nreplica-1)
	s.Writers = make([]*bufio.Writer, nreplica-1)
	s.nextPeer = 0
	s.nPeers = nreplica - 1

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

	go s.ProcessIncoming()
	go s.ProcessOutgoing()

	return s
}

func (s *State) listenToPeer(reader *bufio.Reader) {
	m := message.Message{}
	for {
		err := m.Unmarshal(reader)
		if err != nil {
			// TODO: Reconnect if dies.
			fmt.Println("Read error: ", err)
			return
		}
		s.AddTasks(m)
	}
}

func (s *State) registerConnection(conn net.Conn, i int) {
	s.Connections[i] = conn
	s.Readers[i] = bufio.NewReader(conn)
	s.Writers[i] = bufio.NewWriter(conn)
	go s.listenToPeer(s.Readers[i])
}

func (s *State) WaitForPeers(ln net.Listener) {
	m := &message.Message{}
	for i := 0; i < s.nPeers; i++ {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Bad connection.")
			continue
		}
		s.registerConnection(conn, s.nextPeer)
		m.Unmarshal(s.Readers[s.nextPeer])
		fmt.Printf("%s:%d\n", string(m.Rep.Hostname), m.Rep.Port)
		s.Peers[s.nextPeer] = m.Rep
		s.PeerMap[getkey(m.Rep)] = s.nextPeer
		s.nextPeer++
	}
}

func (s *State) GetPeers(wire net.Conn) {
	buf := bufio.NewReader(wire)
	m := &message.Message{}
	for i := 0; i < s.nPeers; i++ {
		m.Unmarshal(buf)
		fmt.Printf("%s:%d\n", string(m.Rep.Hostname), m.Rep.Port)
		if i > 0 {
			conn, err := net.Dial("tcp",
				fmt.Sprintf("%s:%d", string(m.Rep.Hostname), m.Rep.Port))
			if err != nil {
				fmt.Fprintln(os.Stderr, "Bad connection:", err)
			}
			s.registerConnection(conn, s.nextPeer)
		} else {
			s.Connections[i] = wire
			s.Readers[i] = buf
			s.Writers[i] = bufio.NewWriter(wire)
		}
		s.Peers[i] = m.Rep
		s.PeerMap[getkey(m.Rep)] = i
		s.nextPeer++
	}
	go s.listenToPeer(s.Readers[0])
}

func (s *State) connect(i int) *bufio.Writer {
	host := string(s.Peers[i].Hostname)
	port := s.Peers[i].Port
	conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	return bufio.NewWriter(conn)
}

func (s *State) SendHosts() {
	for i := 0; i < s.nPeers; i++ {
		buf := s.Writers[i]
		m := message.AddHost(string(s.Self.Hostname), s.Self.Port)
		m.Send(buf)
		for j := 0; j < s.nPeers; j++ {
			if i != j {
				m = message.AddHost(string(s.Peers[j].Hostname),
					s.Peers[j].Port)
				m.Send(buf)
			}
		}
	}
}

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
	fmt.Println(i)
	return tasks[:i-1]
}

func (s *State) SendReply(t message.Task) {
	writer := s.Writers[t.HostId]
	m := &message.Message{
		T:     message.PREACCEPTOK,
		Rep:   s.Self,
		Tasks: make([]message.Task, 1),
	}
	m.Tasks[0] = t
	m.Send(writer)
}

func (s *State) SendToAll(tsk []message.Task) {
	for _, w := range s.Writers {
		m := &message.Message{
			T:     message.PREACCEPT,
			Rep:   s.Self,
			Tasks: tsk,
		}
		m.Send(w)
	}
}

func (s *State) ProcessOutgoing() {
	for {
		var tsk []message.Task
		select {
		//case t := <-s.adminTasksOut:
		case t := <-s.okTasksOut:
			fmt.Println("SEND PREACCEPT OK: ", t)
			s.SendReply(t)
			//case t := <-s.commitTasksOut:
			//case t := <-s.acceptTasksOut:
		case t := <-s.preacceptTasksOut:
			tsk = batch(s.preacceptTasksOut, t)
			fmt.Println("SEND PREACCEPT", tsk)
			//case t:= <-s.clientTasksOut:
		}
		s.SendToAll(tsk)
	}
}
