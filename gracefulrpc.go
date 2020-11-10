/*
-- @Time : 2020/11/9 9:55
-- @Author : raoxiaoya
-- @Desc :
*/
package gracefulrpc

import (
	"errors"
	"io"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mars9/codec"
)

const (
	STATUS_RUNNING  uint8 = 1
	STATUS_STOPPING uint8 = 2
	STATUS_STOPED   uint8 = 3

	GRACERPC_KEY = "GRACERPC_RESTART"
)

var hookableSignals = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGTERM,
	syscall.SIGINT,
	//syscall.SIGTSTP, // ctrl+z
}

var defaultDelayTime = 1 * time.Minute
var DefaultWriter = log.New(os.Stdout, "", log.LstdFlags)
var logger Writer

type Writer interface {
	Println(v ...interface{})
	Printf(string, ...interface{})
}

type Config struct {
	DelayTime time.Duration
	Logger    Writer
	CodecType string
}

type gracerpcServer struct {
	listen        net.Listener
	isChild       bool
	sigChan       chan os.Signal
	activeRequest int64
	lock          *sync.Mutex
	status        uint8
	stopTime      time.Time
	config        Config
	isForked      bool
}

func NewServer(gc Config) (server *gracerpcServer) {
	isChild := os.Getenv(GRACERPC_KEY) != ""

	if gc.Logger == nil {
		gc.Logger = DefaultWriter
	}

	server = &gracerpcServer{
		isChild:       isChild,
		sigChan:       make(chan os.Signal),
		activeRequest: 0,
		lock:          &sync.Mutex{},
		status:        STATUS_RUNNING,
		stopTime:      time.Now(),
		config:        gc,
	}

	logger = server.config.Logger

	return
}

func ListenAndServe(network string, address string) error {
	server := NewServer(Config{
		DelayTime: defaultDelayTime,
		Logger:    DefaultWriter,
		CodecType: "json",
	})
	return server.ListenAndServe(network, address)
}

func (server *gracerpcServer) ListenAndServe(network string, address string) error {
	l, err := server.getListener(network, address)
	if err != nil {
		return err
	}
	server.listen = l

	go server.handleSignals()

	if server.isChild {
		syscall.Kill(syscall.Getppid(), syscall.SIGTERM)
	}

	logger.Println("server is running, pid is ", syscall.Getpid())
	server.serve()

	return nil
}

func (server *gracerpcServer) getListener(network string, address string) (l net.Listener, err error) {
	if server.isChild {
		f := os.NewFile(uintptr(3), "")
		l, err = net.FileListener(f)
	} else {
		l, err = net.Listen(network, address)
	}

	return
}

func (server *gracerpcServer) handleSignals() bool {
	var sig os.Signal
	signal.Notify(server.sigChan, hookableSignals...)

	pid := syscall.Getpid()

	for {
		sig = <-server.sigChan
		switch sig {
		case syscall.SIGHUP:
			logger.Println(pid, "Received SIGHUP, restarting...")
			err := server.startProcess()
			if err != nil {
				logger.Println("restarting err: ", err)
			}
		case syscall.SIGINT:
			logger.Println(pid, "Received SIGINT, stop now")
			server.stopNow()
		case syscall.SIGTERM:
			logger.Println(pid, "Received SIGTERM, stop delay")
			server.stopDelay()
		default:
			logger.Printf("Received %v: nothing to do\n", sig)
		}
	}
}

func (server *gracerpcServer) startProcess() (err error) {
	server.lock.Lock()
	defer server.lock.Unlock()

	// only one server instance should fork!
	if server.isForked {
		return errors.New("Another process already forked")
	}

	server.isForked = true

	var files = make([]*os.File, 1)
	tl, ok := server.listen.(*net.TCPListener)
	if !ok {
		return errors.New("get tcp listner file failed.")
	}
	files[0], err = tl.File()
	if err != nil {
		return err
	}

	env := append(
		os.Environ(),
		GRACERPC_KEY+"=1",
	)

	path := os.Args[0]
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = files
	cmd.Env = env

	err = cmd.Start()
	if err != nil {
		logger.Printf("Restart: Failed to launch, error: %v", err)
		return err
	}

	return
}

func (server *gracerpcServer) stopDelay() {
	server.status = STATUS_STOPPING
	err := server.listen.Close()
	if err != nil {
		logger.Println(err)
	}
	server.stopTime = time.Now()
}

func (server *gracerpcServer) stopNow() {
	server.status = STATUS_STOPED
	err := server.listen.Close()
	if err != nil {
		logger.Println(err)
	}
}

func (server *gracerpcServer) serve() {
	flag := true
	for {
		if !flag {
			break
		}
		switch server.status {
		case STATUS_RUNNING:
			flag = server.serveRunning()
		case STATUS_STOPED:
			flag = server.serveStopped()
		case STATUS_STOPPING:
			flag = server.serveStopping()
		}
	}
}

func (server *gracerpcServer) serveRunning() bool {
	conn, e := server.listen.Accept()
	if e != nil {
		return true
	}
	server.activeRequest++
	go func() {
		defer func() {
			server.activeRequest--
		}()

		err := server.serveConn(conn)
		if err != nil {
			logger.Println(err)
			return
		}
	}()

	return true
}

func (server *gracerpcServer) serveConn(conn io.ReadWriteCloser) (err error) {
	codecType := server.config.CodecType
	if codecType == "" {
		codecType = "json"
	}

	switch codecType {
	case "json":
		jsonrpc.ServeConn(conn)
	case "gob":
		rpc.ServeConn(conn)
	case "protobuf":
		rpc.ServeCodec(codec.NewServerCodec(conn))
	default:
		err = errors.New("unknown codecType: " + codecType)
	}

	return
}

func (server *gracerpcServer) serveStopped() bool {
	return false
}

func (server *gracerpcServer) serveStopping() bool {
	if server.activeRequest == 0 {
		return false
	}

	if server.stopTime.Add(server.config.DelayTime).Before(time.Now()) {
		return false
	}

	return true
}
