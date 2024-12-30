package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var (
	imageName     = "shavian-translator"
	inputFileName = defaultEnv("INPUT_FILE", "input")
	inputFileType = defaultEnv("INPUT_FORMAT", "epub")
)

func main() {

	outputFilePath := fmt.Sprintf("/app/output/%s-processed.%s", inputFileName, inputFileType) // Output file in the container
	localOutputPath := fmt.Sprintf("output_local.%s", inputFileType)

	ctx := context.Background()

	// os.Setenv("OLLAMA_PATH", "/Users/csampson/.ollama")
	os.Setenv("INPUT_FILE", inputFileName)
	os.Setenv("INPUT_FORMAT", inputFileType)

	// Step 1: Run Docker Compose
	if err := runDockerCompose(ctx, "compose", "up", "--build", "--wait" /* , "--attach-dependencies"*/); err != nil {
		fmt.Printf("Error running docker-compose: %v\n", err)
		return
	}
	fmt.Println("Docker Compose started successfully.")

	time.Sleep(15 * time.Second) // Wait for a few seconds to ensure the container is up and running

	// Step 2: Wait for containers to complete (if needed)
	// For example, if you want to wait for `go-python-container-instance` to finish:
	if err := waitForContainer(ctx, fmt.Sprintf("%s-instance", imageName)); err != nil {
		fmt.Printf("Error waiting for container: %v\n", err)
		return
	}
	fmt.Println("Containers finished execution.")

	// Step 3: Copy output file from the Go-Python container (if applicable)
	// outputFilePath := "output.txt"
	// localOutputPath := "output_local.txt"
	if err := copyFileFromContainer(fmt.Sprintf("%s-instance", imageName), outputFilePath, localOutputPath); err != nil {
		fmt.Printf("Error copying file from container: %v\n", err)
		return
	}
	fmt.Printf("Output file copied to local system: %s\n", localOutputPath)

	// Step 4: Stop and clean up the Docker Compose setup
	if err := runDockerCompose(ctx, "down"); err != nil {
		fmt.Printf("Error stopping docker-compose: %v\n", err)
		return
	}
	fmt.Println("Docker Compose stopped and cleaned up.")

	// Step 5: Remove docker containers
	if err := removeDockerContainers(ctx); err != nil {
		fmt.Printf("Error removing containers: %v\n", err)
		return
	}
}

// runDockerCompose runs a docker-compose command with the specified arguments.
func runDockerCompose(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "docker", args...)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	// if err := cmd.Run(); err != nil {
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to execute docker compose: %w", err)
	}
	return nil
}

// waitForContainer waits for a specific container to finish execution.
func waitForContainer(ctx context.Context, containerName string) error {
	cmd := exec.CommandContext(ctx, "docker", "wait", containerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to wait for container %s: %w", containerName, err)
	}
	return nil
}

// remove docker containers after the process is done
func removeDockerContainers(ctx context.Context) error {
	// cmd := exec.CommandContext(ctx, "docker", "rm", "-f", fmt.Sprintf("%s-instance", imageName), "ollama-server")
	cmd := exec.CommandContext(ctx, "docker", "rm", "-f", fmt.Sprintf("%s-instance", imageName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}

// copyFileFromContainer copies a file from a container to the local filesystem using `docker cp`.
func copyFileFromContainer(containerName, containerFilePath, localFilePath string) error {
	cmd := exec.Command("docker", "cp", fmt.Sprintf("%s:%s", containerName, containerFilePath), localFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy file from container: %w", err)
	}
	return nil
}

func defaultEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
