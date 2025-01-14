package executors

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
)

// Build builds the docker image from a shuttle plan
func executeDocker(ctx context.Context, context ActionExecutionContext) error {
	dockerFilePath := path.Join(context.ScriptContext.Project.LocalPlanPath, context.Action.Dockerfile)
	projectPath := context.ScriptContext.Project.ProjectPath
	execCmd := exec.Command("docker", "build", "-f", dockerFilePath, projectPath)

	var errStdout, errStderr error
	stdoutIn, _ := execCmd.StdoutPipe()
	stderrIn, _ := execCmd.StderrPipe()
	execCmd.Start()

	go func() {
		_, errStdout = copyAndCapture(os.Stdout, stdoutIn)
	}()

	go func() {
		_, errStderr = copyAndCapture(os.Stderr, stderrIn)
	}()

	err := execCmd.Wait()

	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatalf("failed to capture stdout or stderr\n")
	}
	return nil
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
	// never reached
	panic(true)
	return nil, nil
}
