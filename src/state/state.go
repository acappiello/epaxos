package state

import (
    "bufio"
    "fmt"
    "io"
    "net"
    "os"

    "message"
    "replicainfo"
)

type State struct {
    Self replicainfo.ReplicaInfo
    Peers []replicainfo.ReplicaInfo
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
    s.nextPeer = 0
    s.nPeers = nreplica-1
    return s
}

func (s *State) NewConnection(wire io.Reader) {
    m := &message.Message{}
    m.Unmarshal(wire)
    fmt.Printf("%s:%d\n", string(m.Rep.Hostname), m.Rep.Port)
    s.Peers[s.nextPeer] = m.Rep
    s.nextPeer++
}

func (s *State) connect(i int) *bufio.Writer {
    host := string(s.Peers[i].Hostname)
    port := s.Peers[i].Port
    conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
    return bufio.NewWriter(conn)
}

func (s *State) SendHosts() {
    for i := 0; i < s.nPeers; i++ {
        conn := s.connect(i)
        m := message.AddHost(string(s.Self.Hostname), s.Self.Port)
        m.Marshal(conn)
        conn.Flush()
        for j := 0; j < s.nPeers; j++ {
            if i != j {
                // This is dumb. I just don't want to worry about it now.
                conn := s.connect(i)
                m = message.AddHost(string(s.Peers[j].Hostname),
                    s.Peers[j].Port)
                m.Marshal(conn)
                conn.Flush()
            }
        }
    }
}
