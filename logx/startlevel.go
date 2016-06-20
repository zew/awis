package logx

import "sync"

type startLevel struct {
	lvl              int
	appendStacktrace bool
	sync.Mutex
}

var sl = startLevel{lvl: 2} // one up for this func; one up for Printf
func SL() *startLevel {
	return &sl
}
func (s *startLevel) Incr() *startLevel {
	s.Lock()
	s.lvl++
	s.Unlock()
	return s
}
func (s *startLevel) Decr() *startLevel {
	s.Lock()
	s.lvl--
	s.Unlock()
	return s
}
func (s *startLevel) AppendStacktrace() *startLevel {
	s.Lock()
	s.appendStacktrace = true
	s.Unlock()
	return s
}
