package envite

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type ExecutionMode string

const (
	ExecutionModeStart  ExecutionMode = "start"
	ExecutionModeStop   ExecutionMode = "stop"
	ExecutionModeDaemon ExecutionMode = "daemon"
)

func ParseExecutionMode(value string) (ExecutionMode, error) {
	switch value {
	case "start":
		return ExecutionModeStart, nil
	case "stop":
		return ExecutionModeStop, nil
	case "daemon", "":
		return ExecutionModeDaemon, nil
	}
	return "", ErrInvalidExecutionMode{v: value}
}

func Execute(server *Server, executionMode ExecutionMode) error {
	switch executionMode {
	case ExecutionModeStart:
		return server.env.StartAll(context.Background())
	case ExecutionModeStop:
		err := server.env.StopAll(context.Background())
		if err != nil {
			return err
		}

		return server.env.Cleanup(context.Background())
	case ExecutionModeDaemon:
		go handleDaemonStart(server)
		return server.Start()
	}
	return ErrInvalidExecutionMode{v: string(executionMode)}
}

var asciiArt = `
        ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓                                                                                                         
     ▓▓▓▓▓           ▓▓▓▓    ▓                                                                                                  
   ▓▓▓▓▓▓              ▓▓▓ ▓▓▓                                                                                                  
  ▓▓▓▓▓▓▓▓              ▓▓▓▓▓       ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓      ▓▓▓▓▓▓ ▓▓▓▓▓▓        ▓▓▓▓▓▓  ▓▓  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓▓▓  
 ▓▓▓▓▓▓▓▓▓▓        ▓▓▓▓▓▓▓ ▓▓▓      ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓    ▓▓▓▓▓▓ ▓▓▓▓▓▓▓      ▓▓▓▓▓▓▓  ▓▓  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓▓▓  
▓▓▓▓▓▓▓▓▓▓▓       ▓▓▓▓▓▓    ▓▓      ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓   ▓▓▓▓▓▓  ▓▓▓▓▓▓▓    ▓▓▓▓▓▓▓   ▓▓        ▓▓▓        ▓▓▓           
▓▓▓▓▓▓▓▓▓▓▓    ▓▓▓▓▓▓        ▓▓     ▓▓▓▓▓▓▓          ▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓    ▓▓        ▓▓▓        ▓▓▓           
▓▓▓▓▓▓▓▓▓▓   ▓▓▓▓▓▓▓         ▓▓     ▓▓▓▓▓▓▓▓▓▓▓▓▓▓   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓     ▓▓        ▓▓▓        ▓▓▓▓▓▓▓▓▓▓▓   
▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓         ▓▓     ▓▓▓▓▓▓▓▓▓▓▓▓▓▓   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓    ▓▓▓▓▓▓▓▓▓▓▓▓▓      ▓▓        ▓▓▓        ▓▓▓▓▓▓▓▓▓▓▓   
▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓           ▓▓     ▓▓▓▓▓▓▓          ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓     ▓▓▓▓▓▓▓▓▓▓▓▓      ▓▓        ▓▓▓        ▓▓▓           
▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓             ▓▓      ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓      ▓▓▓▓▓▓▓▓▓▓       ▓▓        ▓▓▓        ▓▓▓           
 ▓▓▓▓▓▓▓▓▓▓▓▓▓▓            ▓▓▓      ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓    ▓▓▓▓▓▓▓       ▓▓▓▓▓▓▓▓        ▓▓        ▓▓▓        ▓▓▓▓▓▓▓▓▓▓▓▓▓ 
  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓          ▓▓▓       ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓     ▓▓▓▓▓▓       ▓▓▓▓▓▓▓         ▓▓        ▓▓▓        ▓▓▓▓▓▓▓▓▓▓▓▓▓ 
   ▓▓▓▓▓▓▓▓▓▓▓▓▓       ▓▓▓▓                                                                                                     
      ▓▓▓▓▓▓▓▓▓▓     ▓▓▓▓                                                                                                       
         ▓▓▓▓▓▓▓▓▓▓▓▓▓`

func handleDaemonStart(server *Server) {
	fmt.Println(asciiArt)
	url := "http://localhost" + server.addr
	open, err := confirmOpenBrowser(url)
	if err != nil {
		fmt.Println("could not confirm open browser window: ", err.Error())
		return
	}

	if !open {
		return
	}

	err = openBrowser(url)
	if err != nil {
		fmt.Println("could not open browser window: ", err.Error())
	}
}

func confirmOpenBrowser(url string) (bool, error) {
	fmt.Println("starting ENVITE daemon server at " + url)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		ticker := time.NewTicker(time.Second)
		i := 10
		for i > 0 {
			select {
			case <-ticker.C:
				fmt.Printf("\rwould you like to open a browser window? [y/N] (%d)", i)
				i--
			case <-ctx.Done():
				return
			}
		}
	}()
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	return strings.ToLower(strings.TrimSpace(response)) == "y", nil
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

type ErrInvalidExecutionMode struct {
	v string
}

func (e ErrInvalidExecutionMode) Error() string {
	return fmt.Sprintf("invalid execution mode %s", e.v)
}
