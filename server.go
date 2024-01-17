package gasp

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Server struct {
	commSocket    string
	httpServer    *http.Server
	handledPaths  []string
	guardedPaths  map[string]func(req *http.Request) (newPath *string)
	eventHandlers map[string][]func(event *ClientEvent)
	eventChan     chan *ServerEvent
	varSetters    map[string]func(req *http.Request) string
	resources     map[string][]byte
	form          *Form
	useTls        bool

	ErrorChan chan error
}

func NewServer(socket string, form ...*Form) (*Server, error) {
	server := Server{}

	err := ValidateSocket(socket)
	if err != nil {
		return nil, err
	}
	server.commSocket = socket

	server.guardedPaths = make(map[string]func(req *http.Request) (newPath *string))
	server.eventHandlers = make(map[string][]func(event *ClientEvent))
	server.eventChan = make(chan *ServerEvent, 10000)
	server.varSetters = make(map[string]func(req *http.Request) string)

	if form != nil && len(form) > 0 {
		server.form = form[0]
	}

	server.ErrorChan = make(chan error)

	return &server, nil
}

func (server *Server) Build() error {
	if len(server.handledPaths) == 0 {
		server.addDefaultHandler()
	}

	server.addWebsocketsHandler()
	return nil
}

func (server *Server) Start() error {
	err := server.Build()
	if err != nil {
		return err
	}

	go func() {
		listener, err := net.Listen("tcp", server.commSocket)
		if err != nil {
			server.sendError(err)
			return
		}
		defer func(listener net.Listener) {
			err := listener.Close()
			if err != nil {
				server.sendError(err)
				return
			}
		}(listener)

		server.httpServer = &http.Server{}
		err = server.httpServer.Serve(listener)
		if err != nil {
			server.sendError(err)
			server.httpServer = nil
			return
		}
	}()

	time.Sleep(time.Second)
	return nil
}

func (server *Server) StartWithTLS(certFile string, keyFile string) error {
	server.useTls = true

	err := server.Build()
	if err != nil {
		return err
	}

	go func() {
		listener, err := net.Listen("tcp", server.commSocket)
		if err != nil {
			server.sendError(err)
			return
		}
		defer func(listener net.Listener) {
			err := listener.Close()
			if err != nil {
				server.sendError(err)
				return
			}
		}(listener)

		server.httpServer = &http.Server{}
		err = server.httpServer.ServeTLS(listener, certFile, keyFile)
		if err != nil {
			server.sendError(err)
			server.httpServer = nil
			return
		}
	}()

	time.Sleep(time.Second)
	return nil
}

func (server *Server) Stop() error {
	if server.httpServer != nil {
		err := server.httpServer.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (server *Server) SendEvent(event *ServerEvent) {
	select {
	case server.eventChan <- event:
	default:
	}
}

func (server *Server) AddResources(directory string, excludedFileExtensions ...string) error {
	if strings.TrimSpace(directory) == "" {
		return errors.New("parameter 'directory' cannot be empty/whitespace")
	}

	dirInfo, err := os.Stat(directory)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("resources directory '%s' does not exist", directory)
	}

	if !dirInfo.IsDir() {
		return fmt.Errorf("'%s' is not a directory", directory)
	}

	if directory[len(directory)-1:] != "/" {
		directory += "/"
	}

	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			for _, ext := range excludedFileExtensions {
				if strings.HasSuffix(info.Name(), ext) {
					return nil
				}
			}
			resourceContents, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			urlPath := strings.Replace(path, directory, "", 1)
			return server.AddRouteHandler(urlPath, func(rw http.ResponseWriter, req *http.Request) {
				_, err := rw.Write(resourceContents)
				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					server.ErrorChan <- err
				}
			})
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to add resources directory: %v", err)
	}

	return nil
}

func (server *Server) AddRouteHandler(path string, handler func(http.ResponseWriter, *http.Request)) error {
	updatedPath, err := server.validateNewPath(path)
	if err != nil {
		return err
	}

	http.HandleFunc(updatedPath, handler)
	server.handledPaths = append(server.handledPaths, updatedPath)
	return nil
}

func (server *Server) AddView(path string, html string) error {
	updatedPath, err := server.validateNewPath(path)
	if err != nil {
		return err
	}

	vars := map[string]string{"server_socket": server.commSocket}
	handler := func(rw http.ResponseWriter, req *http.Request) {
		if guardFunc, ok := server.guardedPaths[updatedPath]; ok {
			newPath := guardFunc(req)
			if newPath != nil {
				if (*newPath)[:1] != "/" {
					*newPath = "/" + *newPath
				}
				http.Redirect(rw, req, "http://"+server.commSocket+*newPath, http.StatusFound)
			}
		}

		handledEvents := make([]string, 0)
		for event := range server.eventHandlers {
			handledEvents = append(handledEvents, event)
		}

		for varName, setter := range server.varSetters {
			if strings.Contains(html, "<!--"+varName+"-->") {
				vars[varName] = setter(req)
			}
		}

		view := GenerateView(html, vars, handledEvents, server.useTls)
		_, err := rw.Write([]byte(view))
		if err != nil {
			server.sendError(err)
		}
	}

	http.HandleFunc(updatedPath, handler)
	server.handledPaths = append(server.handledPaths, updatedPath)
	return nil
}

func (server *Server) AddEventHandler(view string, elementId string, eventType string, handler func(event *ClientEvent)) {
	eventName := fmt.Sprintf("%s#%s!%s", view, elementId, eventType)
	if handlers, ok := server.eventHandlers[eventName]; ok {
		handlers = append(handlers, handler)
		server.eventHandlers[eventName] = handlers
	} else {
		handlers := make([]func(event *ClientEvent), 1)
		handlers[0] = handler
		server.eventHandlers[eventName] = handlers
	}
}

func (server *Server) AddVariableSetter(variableName string, setter func(req *http.Request) string) error {
	reservedNames := []string{"gasp_css", "gasp_js", "server_socket", "now"}
	for _, name := range reservedNames {
		if strings.HasPrefix(variableName, name) {
			return fmt.Errorf("variable name '%s' is reserved", variableName)
		}
	}
	server.varSetters[variableName] = setter
	return nil
}

func (server *Server) AddRouteGuard(path string, guardFunc func(req *http.Request) (newPath *string)) error {
	path = strings.TrimSpace(path)

	if path == "" {
		path = "/"
	}

	if path[:1] != "/" {
		path = "/" + path
	}

	if path == "/gaspws" {
		return errors.New("path cannot be '/gaspws', which is reserved for the WebSockets channel")
	}

	if _, ok := server.guardedPaths[path]; ok {
		return fmt.Errorf("path '%s' already guarded", path)
	}

	server.guardedPaths[path] = guardFunc
	return nil
}

func (server *Server) sendError(err error) {
	select {
	case server.ErrorChan <- err:
	default:
	}
}

func (server *Server) validateNewPath(path string) (string, error) {
	path = strings.TrimSpace(path)

	if path == "" {
		path = "/"
	}

	if path[:1] != "/" {
		path = "/" + path
	}

	if path == "/gaspws" {
		return "", errors.New("path cannot be '/gaspws', which is reserved for the WebSockets channel")
	}

	for _, p := range server.handledPaths {
		if p == path {
			return "", fmt.Errorf("path '%s' already handled", path)
		}
	}

	return path, nil
}

func (server *Server) addDefaultHandler() {
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		_, err := rw.Write([]byte("Gasp Server Online"))
		if err != nil {
			server.sendError(err)
		}
	})
	server.handledPaths = append(server.handledPaths, "/")
}

func (server *Server) processOutgoingEvents(ws *websocket.Conn, abortChan <-chan bool) {
	for {
		select {
		case <-abortChan:
			return
		case outEvt := <-server.eventChan:
			evtBytes, err := json.Marshal(outEvt)
			if err != nil {
				server.sendError(err)
			}
			err = ws.WriteMessage(websocket.TextMessage, evtBytes)
			if err != nil {
				server.sendError(err)
			}
		}
	}
}

func (server *Server) processIncomingEvents(ws *websocket.Conn) {
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			server.sendError(err)
			return
		}

		event := ClientEvent{}
		err = json.Unmarshal(p, &event)
		if err != nil {
			server.sendError(err)
			return
		}

		if server.form != nil {
			event.Form = server.form
		}

		if eventHandlers, ok := server.eventHandlers[event.View+"#"+event.Id+"!"+event.Type]; ok {
			for _, handler := range eventHandlers {
				handler(&event)
			}
		}

		if eventHandlers, ok := server.eventHandlers["*#"+event.Id+"!"+event.Type]; ok {
			for _, handler := range eventHandlers {
				handler(&event)
			}
		}

		if eventHandlers, ok := server.eventHandlers[event.View+"#*!"+event.Type]; ok {
			for _, handler := range eventHandlers {
				handler(&event)
			}
		}

		if eventHandlers, ok := server.eventHandlers["*#*!"+event.Type]; ok {
			for _, handler := range eventHandlers {
				handler(&event)
			}
		}
	}
}

func (server *Server) getWebsocketsHandler(path string) func(rw http.ResponseWriter, req *http.Request) {
	for _, p := range server.handledPaths {
		if p == path {
			return nil
		}
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return func(rw http.ResponseWriter, req *http.Request) {
		ws, err := upgrader.Upgrade(rw, req, nil)
		if err != nil {
			server.sendError(err)
		}

		server.SendEvent(&ServerEvent{
			Type: "ws_info",
			Text: "client connected",
		})

		abortChan := make(chan bool)
		go server.processOutgoingEvents(ws, abortChan)
		defer func() {
			abortChan <- true
		}()

		server.processIncomingEvents(ws)
	}
}

func (server *Server) addWebsocketsHandler() {
	path := "/gaspws"
	handler := server.getWebsocketsHandler(path)
	if handler == nil {
		return
	}
	http.HandleFunc(path, handler)
	server.handledPaths = append(server.handledPaths, path)
}
