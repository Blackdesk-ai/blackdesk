package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	DefaultAppName      = "blackdesk"
	DefaultRepo         = "Blackdesk-ai/blackdesk"
	defaultAPIBaseURL   = "https://api.github.com"
	defaultDownloadBase = "https://github.com"
)

type Config struct {
	AppName          string
	Repo             string
	APIBaseURL       string
	DownloadsBaseURL string
	HTTPClient       *http.Client
}

type Client struct {
	appName          string
	repo             string
	apiBaseURL       string
	downloadsBaseURL string
	httpClient       *http.Client
}

type CheckResult struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	Comparable      bool
}

type UpgradeResult struct {
	PreviousVersion  string
	InstalledVersion string
	AssetName        string
	ExecutablePath   string
	RestartRequired  bool
	AlreadyCurrent   bool
}

func New(config Config) *Client {
	appName := strings.TrimSpace(config.AppName)
	if appName == "" {
		appName = DefaultAppName
	}
	repo := strings.TrimSpace(config.Repo)
	if repo == "" {
		repo = DefaultRepo
	}
	apiBaseURL := strings.TrimRight(strings.TrimSpace(config.APIBaseURL), "/")
	if apiBaseURL == "" {
		apiBaseURL = defaultAPIBaseURL
	}
	downloadsBaseURL := strings.TrimRight(strings.TrimSpace(config.DownloadsBaseURL), "/")
	if downloadsBaseURL == "" {
		downloadsBaseURL = defaultDownloadBase
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{
		appName:          appName,
		repo:             repo,
		apiBaseURL:       apiBaseURL,
		downloadsBaseURL: downloadsBaseURL,
		httpClient:       httpClient,
	}
}

func Default() *Client {
	return New(Config{})
}

func (c *Client) Check(ctx context.Context, currentVersion string) (CheckResult, error) {
	currentVersion = normalizeVersion(currentVersion)
	latestVersion, err := c.LatestVersion(ctx)
	if err != nil {
		return CheckResult{}, err
	}
	result := CheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
	}
	comparison, comparable := compareVersions(currentVersion, latestVersion)
	result.Comparable = comparable
	result.UpdateAvailable = comparable && comparison < 0
	if currentVersion == "dev" && latestVersion != "" {
		result.UpdateAvailable = true
	}
	return result, nil
}

func (c *Client) LatestVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+"/repos/"+c.repo+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", c.appName+"/updater")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("latest release lookup failed: %s", strings.TrimSpace(string(body)))
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	version := normalizeVersion(payload.TagName)
	if version == "" {
		return "", errors.New("latest release did not include a tag")
	}
	return version, nil
}

func (c *Client) Upgrade(ctx context.Context, executablePath, currentVersion, requestedVersion string) (UpgradeResult, error) {
	targetVersion, err := c.resolveTargetVersion(ctx, requestedVersion)
	if err != nil {
		return UpgradeResult{}, err
	}

	result := UpgradeResult{
		PreviousVersion:  normalizeVersion(currentVersion),
		InstalledVersion: targetVersion,
		ExecutablePath:   executablePath,
	}
	if comparison, comparable := compareVersions(result.PreviousVersion, targetVersion); comparable && comparison == 0 {
		result.AlreadyCurrent = true
		return result, nil
	}

	osName, arch, err := detectPlatform()
	if err != nil {
		return UpgradeResult{}, err
	}
	assetName := c.archiveName(targetVersion, osName, arch)
	extractedBinary, err := c.downloadReleaseBinary(ctx, targetVersion, assetName)
	if err != nil {
		return UpgradeResult{}, err
	}
	defer os.Remove(extractedBinary)

	if err := installBinary(extractedBinary, executablePath); err != nil {
		return UpgradeResult{}, err
	}

	result.AssetName = assetName
	result.RestartRequired = runtime.GOOS == "windows"
	return result, nil
}

func (c *Client) resolveTargetVersion(ctx context.Context, requestedVersion string) (string, error) {
	requestedVersion = normalizeVersion(requestedVersion)
	if requestedVersion == "" || requestedVersion == "latest" {
		return c.LatestVersion(ctx)
	}
	return requestedVersion, nil
}

func (c *Client) archiveName(version, osName, arch string) string {
	if osName == "windows" {
		return fmt.Sprintf("%s_%s_%s_%s.zip", c.appName, version, osName, arch)
	}
	return fmt.Sprintf("%s_%s_%s_%s.tar.gz", c.appName, version, osName, arch)
}

func (c *Client) checksumName(version string) string {
	return fmt.Sprintf("%s_%s_SHA256SUMS.txt", c.appName, version)
}

func (c *Client) releaseBaseURL(version string) string {
	return fmt.Sprintf("%s/%s/releases/download/v%s", c.downloadsBaseURL, c.repo, version)
}

func (c *Client) downloadReleaseBinary(ctx context.Context, version, assetName string) (string, error) {
	tmpDir, err := os.MkdirTemp("", c.appName+"-upgrade-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, assetName)
	checksumName := c.checksumName(version)
	checksumPath := filepath.Join(tmpDir, checksumName)
	baseURL := c.releaseBaseURL(version)

	if err := c.downloadFile(ctx, baseURL+"/"+assetName, archivePath); err != nil {
		return "", err
	}
	if err := c.downloadFile(ctx, baseURL+"/"+checksumName, checksumPath); err != nil {
		return "", err
	}
	if err := verifyChecksum(archivePath, checksumPath, assetName); err != nil {
		return "", err
	}

	extractedPath := filepath.Join(tmpDir, binaryName())
	if runtime.GOOS == "windows" {
		extractedPath += ".exe"
	}
	if strings.HasSuffix(assetName, ".zip") {
		if err := extractZipBinary(archivePath, extractedPath); err != nil {
			return "", err
		}
	} else {
		if err := extractTarGzBinary(archivePath, extractedPath); err != nil {
			return "", err
		}
	}

	finalFile, err := os.CreateTemp("", c.appName+"-upgrade-binary-*")
	if err != nil {
		return "", err
	}
	finalPath := finalFile.Name()
	if err := finalFile.Close(); err != nil {
		return "", err
	}
	if err := copyFile(extractedPath, finalPath, 0o755); err != nil {
		return "", err
	}
	return finalPath, nil
}

func (c *Client) downloadFile(ctx context.Context, sourceURL, destination string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.appName+"/updater")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("download failed for %s: %s", sourceURL, strings.TrimSpace(string(body)))
	}

	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func verifyChecksum(archivePath, checksumPath, assetName string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	actual := hex.EncodeToString(hash.Sum(nil))

	sums, err := os.ReadFile(checksumPath)
	if err != nil {
		return err
	}
	expected := ""
	for _, line := range strings.Split(string(sums), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if fields[1] == assetName {
			expected = fields[0]
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("checksum missing for %s", assetName)
	}
	if !strings.EqualFold(expected, actual) {
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}
	return nil
}

func extractTarGzBinary(archivePath, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()

	reader := tar.NewReader(gz)
	target := filepath.Base(destination)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if header.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(header.Name) != target {
			continue
		}
		out, err := os.Create(destination)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, reader); err != nil {
			out.Close()
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}
		return os.Chmod(destination, 0o755)
	}
	return fmt.Errorf("archive did not contain %s", target)
}

func extractZipBinary(archivePath, destination string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	target := filepath.Base(destination)
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(file.Name) != target {
			continue
		}
		in, err := file.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(destination)
		if err != nil {
			in.Close()
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			in.Close()
			out.Close()
			return err
		}
		in.Close()
		if err := out.Close(); err != nil {
			return err
		}
		return os.Chmod(destination, 0o755)
	}
	return fmt.Errorf("archive did not contain %s", target)
}

func installBinary(sourcePath, executablePath string) error {
	if runtime.GOOS == "windows" {
		return stageWindowsReplacement(sourcePath, executablePath)
	}
	tempPath := executablePath + ".tmp"
	if err := copyFile(sourcePath, tempPath, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tempPath, executablePath); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return nil
}

func stageWindowsReplacement(sourcePath, executablePath string) error {
	stagePath := executablePath + ".new"
	if err := copyFile(sourcePath, stagePath, 0o755); err != nil {
		return err
	}

	scriptFile, err := os.CreateTemp("", "blackdesk-upgrade-*.cmd")
	if err != nil {
		return err
	}
	scriptPath := scriptFile.Name()
	script := windowsReplaceScript(stagePath, executablePath)
	if _, err := scriptFile.WriteString(script); err != nil {
		scriptFile.Close()
		return err
	}
	if err := scriptFile.Close(); err != nil {
		return err
	}

	cmd := exec.Command("cmd", "/C", "start", "", "/B", scriptPath)
	return cmd.Start()
}

func windowsReplaceScript(sourcePath, executablePath string) string {
	quotedSource := strings.ReplaceAll(sourcePath, `"`, `""`)
	quotedTarget := strings.ReplaceAll(executablePath, `"`, `""`)
	return "@echo off\r\n" +
		"setlocal\r\n" +
		"for /L %%i in (1,1,10) do (\r\n" +
		"  move /Y \"" + quotedSource + "\" \"" + quotedTarget + "\" > nul 2>&1 && goto done\r\n" +
		"  ping 127.0.0.1 -n 2 > nul\r\n" +
		")\r\n" +
		":done\r\n" +
		"del \"%~f0\"\r\n"
}

func copyFile(sourcePath, destinationPath string, mode os.FileMode) error {
	in, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Chmod(destinationPath, mode)
}

func detectPlatform() (string, string, error) {
	osName := runtime.GOOS
	switch osName {
	case "darwin", "linux", "windows":
	default:
		return "", "", fmt.Errorf("unsupported operating system: %s", osName)
	}

	arch := runtime.GOARCH
	switch arch {
	case "amd64", "arm64":
	default:
		return "", "", fmt.Errorf("unsupported architecture: %s", arch)
	}

	if osName == "darwin" && arch == "amd64" && runningUnderRosetta() {
		arch = "arm64"
	}
	return osName, arch, nil
}

func runningUnderRosetta() bool {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "amd64" {
		return false
	}
	output, err := exec.Command("sysctl", "-n", "sysctl.proc_translated").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "1"
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return DefaultAppName
	}
	return DefaultAppName
}
