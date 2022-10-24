package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

var vaultClient = getVaultClient()

func main() {
	pathArg := os.Args[1]
	mount, secretPath, _ := strings.Cut(pathArg, "/")
	secretVersion, err := fetchVersion(mount, secretPath)
	if err != nil {
		log.Fatalf("failed to fetch secret metadata: %v", err)
	}
	fileName := fmt.Sprintf("%d__-__%s", secretVersion, hashFilename(secretPath))
	filePath, err := filePath()
	if err != nil {
		log.Fatalf("failed to create memory path: %v", err)
	}
	_, err = os.Stat(fmt.Sprintf("%s/%s", filePath, fileName))
	if os.IsNotExist(err) {
		data, err := fetchEnvironment(mount, secretPath)
		if err != nil {
			log.Fatalf("failed to fetch secret: %v", err)
		}
		err = writeFile(data, filePath, fileName)
		if err != nil {
			log.Fatalf("failed to write to memory: %v", err)
		}
	}
	fmt.Printf("%s/%s", filePath, fileName)
}

func writeFile(data map[string]interface{}, filePath, fileName string) error {
	file, err := os.Create(fmt.Sprintf("%s/%s", filePath, fileName))
	if err != nil {
		return fmt.Errorf("unable to create file: %w", err)
	}
	defer file.Close()

	for key, value := range data {
		file.WriteString(fmt.Sprintf("%s=\"%s\"\n", key, value))
	}
	return err
}

func fetchVersion(mount, path string) (int, error) {
	data, err := vaultClient.KVv2(mount).GetMetadata(context.Background(), path)
	if err != nil {
		return 0, fmt.Errorf("unable to read metadata: %w", err)
	}
	return data.CurrentVersion, err
}

func fetchEnvironment(mount, path string) (map[string]interface{}, error) {
	data, err := vaultClient.KVv2(mount).Get(context.Background(), path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %w", err)
	}
	return data.Data, err
}

func hashFilename(s string) string {
	hasher := sha256.New()
	hasher.Write([]byte(s))
	name := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return name
}

func filePath() (string, error) {
	macosPath := "/Volumes/vaultenv"
	linuxPath := "/dev/shm"

	if runtime.GOOS == "darwin" {
		err := handleMacos(macosPath)
		return macosPath, err
	} else {
		return linuxPath, nil
	}

}

func handleMacos(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := createRamDisk()
		return err
	}
	return err
}

func createRamDisk() error {
	cmd := exec.Command("hdiutil", "attach", "-nomount", "ram://40960") // desired size in MiB * 2048
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	devicePath := string(bytes.TrimSpace(out))

	cmd = exec.Command("diskutil", "partitionDisk", devicePath, "1", "GPTFormat", "APFS", "vaultenv", "'100%'")
	out, err = cmd.Output()
	if err != nil {
		return err
	}
	return err
}

func getVaultClient() *vault.Client {
	vaultClient, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %w", err)
	}

	token, err := vaultToken()
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %w", err)
	}

	vaultClient.SetToken(token)
	return vaultClient
}

func vaultToken() (string, error) {
	home := os.Getenv("HOME")
	content, err := ioutil.ReadFile(fmt.Sprintf("%s/.vault-token", home))
	if err != nil {
		return "", fmt.Errorf("unable to get vault token, are you signed in?")
	}
	return string(content), nil
}
