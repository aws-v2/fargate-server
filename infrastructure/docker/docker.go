package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"mini-fargate/infrastructure/models"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"

	"mini-fargate/logger"
)

func RunContainer(req models.TaskRequest) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	defer cli.Close()

	ctx := context.Background()
	reader, err := cli.ImagePull(ctx, req.Image, image.PullOptions{})
	if err == nil {
		defer reader.Close()
		io.Copy(os.Stdout, reader)
	}

	var env []string
	for k, v := range req.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for containerPortStr, hostPort := range req.Ports {
		cp := nat.Port(fmt.Sprintf("%s/tcp", containerPortStr))
		portBindings[cp] = []nat.PortBinding{{HostPort: fmt.Sprintf("%d", hostPort)}}
		exposedPorts[cp] = struct{}{}
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        req.Image,
		Env:          env,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{PortBindings: portBindings}, nil, nil, req.Name)

	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

// TaskStatusCallback is used to send status updates back to NATS
type TaskStatusCallback func(status, message string, result *models.NATSResponse)

func RunLambdaContainer(ctx context.Context, inv models.NATSInvocation, callback TaskStatusCallback) (string, string, int, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Log.Error("Error creating docker client", zap.Error(err))
		return "", "", -1, err
	}
	defer cli.Close()

	imageName := inv.Image
	if imageName == "" {
		imageName = "golang:1.22-alpine"
	}

	// 1. Check if image exists, if not send "inprogres" for container
	_, _, err = cli.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		callback("task.status.container.inprogres", "Downloading container image: "+imageName, nil)

		reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
		if err != nil {
			logger.Log.Error("Error pulling image", zap.String("image", imageName), zap.Error(err))
			return "", "", -1, fmt.Errorf("failed to pull image: %w", err)
		}
		defer reader.Close()
		io.Copy(os.Stdout, reader)

		callback("task.status.container.complete", "Container image downloaded: "+imageName, nil)
	}

	// 2. Start task in progress
	callback("task.status.inprogres", "Preparing container for execution", nil)

	// Prepare Env
	var env []string
	for k, v := range inv.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	payloadJSON, _ := json.Marshal(inv.Payload)
	env = append(env, fmt.Sprintf("PAYLOAD=%s", string(payloadJSON)))

	// Resources
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:   inv.Resources.Memory * 1024 * 1024, // MB to Bytes
			NanoCPUs: int64(inv.Resources.CPU * 1e9),     // CPU to NanoCPUs
		},
		NetworkMode: "none",
	}

	// Mount artifact path read-only
	if inv.Execution.Path != "" {
		hostConfig.Mounts = []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   inv.Execution.Path,
				Target:   "/var/task",
				ReadOnly: true,
			},
		}
	}

	// Determine command
	var cmd []string

	// 1. If explicit command is provided in invocation, use it (Generic Support for Java, Python, Rust, etc.)
	if len(inv.Execution.Command) > 0 {
		cmd = inv.Execution.Command
		logger.Log.Info("Using explicit command", zap.String("task_id", inv.TaskID), zap.Strings("cmd", cmd))
	} else {
		// 2. Legacy/Fallback Logic: Infer based on Kind or Image
		executionKind := inv.Execution.Kind
		if executionKind == "binary" {
			if strings.Contains(imageName, "node") || strings.Contains(imageName, "javascript") {
				executionKind = "node"
			} else if strings.Contains(imageName, "java") || strings.Contains(imageName, "jdk") || strings.Contains(imageName, "openjdk") || strings.Contains(imageName, "corretto") {
				executionKind = "java"
			} else if strings.Contains(imageName, "python") {
				executionKind = "python"
			}
		}

		logger.Log.Info("Inferred Execution Kind",
			zap.String("task_id", inv.TaskID),
			zap.String("kind", inv.Execution.Kind),
			zap.String("inferred", executionKind),
			zap.String("image", imageName),
		)

		switch executionKind {
		case "image":
			// Pure Docker Image: use the image's default CMD/ENTRYPOINT
			cmd = nil
		case "binary":
			// Copy to /tmp to handle Windows mount permissions (chmod +x)
			cmd = []string{"sh", "-c", "if [ -d /var/task ]; then if [ -f /var/task/handler ]; then cp /var/task/handler /tmp/executor; elif [ -f /var/task/bootstrap ]; then cp /var/task/bootstrap /tmp/executor; else echo 'Error: /var/task is a directory but no handler or bootstrap found'; ls -la /var/task; exit 1; fi; else cp /var/task /tmp/executor; fi && chmod +x /tmp/executor && /tmp/executor"}
		case "node", "nodejs":
			// Node checks for multiple potential entry files
			cmd = []string{"sh", "-c", "if [ -d /var/task ]; then if [ -f /var/task/index.js ]; then node /var/task/index.js; elif [ -f /var/task/handler.js ]; then node /var/task/handler.js; elif [ -f /var/task/handler ]; then node /var/task/handler; else echo 'Error: /var/task is a directory but no index.js, handler.js or handler found'; ls -la /var/task; exit 1; fi; else node /var/task; fi"}
		case "java":
			// Java looks for a JAR file
			// Check if /var/task is a directory or a file
			cmd = []string{"sh", "-c", "if [ -d /var/task ]; then JAR=$(ls /var/task/*.jar | head -n 1); if [ -n \"$JAR\" ]; then java -jar \"$JAR\"; else echo 'Error: No .jar file found in /var/task'; ls -la /var/task; exit 1; fi; else cp /var/task /tmp/function.jar && java -jar /tmp/function.jar; fi"}
		case "python":
			// Python looks for main.py, app.py, or lambda_function.py
			cmd = []string{"sh", "-c", "if [ -d /var/task ]; then if [ -f /var/task/main.py ]; then python /var/task/main.py; elif [ -f /var/task/app.py ]; then python /var/task/app.py; elif [ -f /var/task/lambda_function.py ]; then python /var/task/lambda_function.py; else echo 'Error: /var/task is a directory but no main.py, app.py or lambda_function.py found'; ls -la /var/task; exit 1; fi; else python /var/task; fi"}
		default:
			// Default fallback
			cmd = []string{"sh", "-c", "if [ -f /var/task ]; then cp /var/task /tmp/executor && chmod +x /tmp/executor && /tmp/executor; else echo 'No executable found at /var/task'; exit 1; fi"}
		}
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Env:   env,
		Cmd:   cmd,
	}, hostConfig, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}, nil, "lambda-"+inv.TaskID)

	if err != nil {
		logger.Log.Error("Error creating container", zap.Error(err))
		return "", "", -1, err
	}
	defer cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true})

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		logger.Log.Error("Error starting container", zap.Error(err))
		return "", "", -1, err
	}

	// Wait for container to exit
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	var exitCode int
	select {
	case err := <-errCh:
		if err != nil {
			logger.Log.Error("Error waiting for container", zap.Error(err))
			return "", "", -1, err
		}
	case status := <-statusCh:
		exitCode = int(status.StatusCode)
	case <-ctx.Done():
		logger.Log.Warn("Context cancelled while waiting for container", zap.Error(ctx.Err()))
		return "", "", -1, ctx.Err()
	}

	// Capture logs
	logOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}
	logs, err := cli.ContainerLogs(context.Background(), resp.ID, logOptions)
	if err != nil {
		logger.Log.Error("Error fetching logs", zap.Error(err))
		return "", "", exitCode, err
	}
	defer logs.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	_, err = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, logs)
	if err != nil {
		logger.Log.Error("Error copying logs", zap.Error(err))
		return "", "", exitCode, err
	}

	// Note: task.status.complete is now handled in nats.go to avoid duplication and ensure it has the result.
	return stdoutBuf.String(), stderrBuf.String(), exitCode, nil
}
