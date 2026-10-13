//go:build js

package game

// TextServer is a no-op in the browser (no IPC sockets).
type TextServer struct{}

func NewTextServer(string, int) *TextServer { return &TextServer{} }
func (s *TextServer) Write(string)          {}
