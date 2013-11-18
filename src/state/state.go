package state

import (
    "bufio"
    "fmt"
    "net"
    "os"

    "message"
    "replicainfo"
)

type State struct {
    Self replicainfo.ReplicaInfo
    Peers []replicainfo.ReplicaInfo
    Connections []net.Conn
    Readers []*bufio.Reader
    Writers []*bufio.Writer
    nextPeer int
    nPeers int
}

func Initialize(port int, nreplica int) (*State) {
    s := new(State)
    // TODO: Handle error.
    host, _ := os.Hostname()
    s.Self.Hostname = []byte(host)
    s.Self.Port = port
    s.Peers = make([]replicainfo.ReplicaInfo, nreplica-1)
    s.Connections = make([]net.Conn, nreplica-1)
    s.Readers = make([]*bufio.Reader, nreplica-1)
    s.Writers = make([]*bufio.Writer, nreplica-1)
    s.nextPeer = 0
    s.nPeers = nreplica-1
    return s
}

func (s *State) registerConnection(conn net.Conn, i int) {
    s.Connections[i] = conn
    s.Readers[i] = bufio.NewReader(conn)
    s.Writers[i] = bufio.NewWriter(conn)
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
        s.nextPeer++
    }
}

func (s *State) GetPeers(wire net.Conn) {
    buf := bufio.NewReader(wire)
    m := &message.Message{}
    for i := 0; i < s.nPeers; i++ {
        m.Unmarshal(buf)
        fmt.Printf("%s:%d\n", string(m.Rep.Hostname), m.Rep.Port)
        conn, err := net.Dial("tcp",
            fmt.Sprintf("%s:%d", string(m.Rep.Hostname), m.Rep.Port))
        if err != nil {
            fmt.Fprintln(os.Stderr, "Bad connection:", err)
        }
        s.Peers[i] = m.Rep
        s.registerConnection(conn, s.nextPeer)
        s.nextPeer++
    }
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
