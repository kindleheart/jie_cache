package app

import (
	"github.com/gin-gonic/gin"
	"jie_cache/api"
	"jie_cache/consistenthash"
	"jie_cache/peer"
	"log"
	"sync"
)

const (
	defaultReplicas = 50
	basePath        = "/jie_cache"
)

type Server struct {
	host        string
	engine      *gin.Engine
	apiRouter   *api.Router
	mu          sync.Mutex
	consistent  *consistenthash.Consistent
	httpGetters map[string]peer.PeerGetter
}

func NewServer(mode, host string) *Server {
	return &Server{
		host:      host,
		engine:    NewGinEngine(mode),
		apiRouter: api.NewRouter(),
	}
}

func (s *Server) Start() {
	s.apiRouter.SetupRouter(s.engine)
	s.engine.Run(s.host)
}

func (s *Server) Set(nodes ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consistent = consistenthash.New(defaultReplicas, nil)
	s.consistent.Add(nodes...)
	s.httpGetters = make(map[string]peer.PeerGetter, len(nodes))
	for _, node := range nodes {
		s.httpGetters[node] = peer.NewHttpGetter(node + basePath)
	}
}

// PickPeer picks a peer according to key
func (s *Server) PickPeer(key string) (peer.PeerGetter, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if node := s.consistent.Get(key); node != "" && node != s.host {
		log.Printf("Pick node %s", node)
		return s.httpGetters[node], true
	}
	return nil, false
}

var _ peer.PeerPicker = (*Server)(nil)
