package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCheckDetectsAvailableUpdate(t *testing.T) {
	client := New(Config{
		APIBaseURL: "https://api.example.test",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/repos/Blackdesk-ai/blackdesk/releases/latest" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}
			return jsonResponse(`{"tag_name":"v0.2.0"}`), nil
		})},
	})

	result, err := client.Check(context.Background(), "0.1.0")
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}
	if !result.UpdateAvailable {
		t.Fatal("expected update to be available")
	}
	if result.LatestVersion != "0.2.0" {
		t.Fatalf("expected latest version to be normalized, got %q", result.LatestVersion)
	}
}

func TestCheckSkipsComparisonForDevBuilds(t *testing.T) {
	client := New(Config{
		APIBaseURL: "https://api.example.test",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(`{"tag_name":"v0.2.0"}`), nil
		})},
	})

	result, err := client.Check(context.Background(), "dev")
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}
	if result.Comparable {
		t.Fatal("expected dev build to skip version comparison")
	}
	if !result.UpdateAvailable {
		t.Fatal("expected dev build to advertise the latest published release")
	}
}

func TestUpgradeDownloadsVerifiesAndInstallsRelease(t *testing.T) {
	version := "0.1.0"
	assetName := fmt.Sprintf("blackdesk_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName = fmt.Sprintf("blackdesk_%s_%s_%s.zip", version, runtime.GOOS, runtime.GOARCH)
	}
	binaryName := "blackdesk"
	if runtime.GOOS == "windows" {
		binaryName = "blackdesk.exe"
	}

	archiveBytes := buildArchive(t, assetName, binaryName, []byte("new-binary"))
	checksumBytes := buildChecksums(assetName, archiveBytes)
	execDir := t.TempDir()
	executablePath := filepath.Join(execDir, binaryName)
	if err := os.WriteFile(executablePath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}

	client := New(Config{
		APIBaseURL:       "https://api.example.test",
		DownloadsBaseURL: "https://downloads.example.test",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.String() {
			case "https://api.example.test/repos/Blackdesk-ai/blackdesk/releases/latest":
				return jsonResponse(`{"tag_name":"v0.1.0"}`), nil
			case "https://downloads.example.test/Blackdesk-ai/blackdesk/releases/download/v0.1.0/" + assetName:
				return bytesResponse(archiveBytes), nil
			case "https://downloads.example.test/Blackdesk-ai/blackdesk/releases/download/v0.1.0/blackdesk_0.1.0_SHA256SUMS.txt":
				return bytesResponse(checksumBytes), nil
			default:
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("not found")),
					Header:     make(http.Header),
				}, nil
			}
		})},
	})

	result, err := client.Upgrade(context.Background(), executablePath, "0.0.9", "")
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if result.AssetName != assetName {
		t.Fatalf("unexpected asset name: %q", result.AssetName)
	}

	if runtime.GOOS == "windows" {
		stagedPath := executablePath + ".new"
		data, err := os.ReadFile(stagedPath)
		if err != nil {
			t.Fatalf("expected staged binary on windows: %v", err)
		}
		if string(data) != "new-binary" {
			t.Fatalf("unexpected staged binary contents: %q", string(data))
		}
		return
	}

	data, err := os.ReadFile(executablePath)
	if err != nil {
		t.Fatalf("read upgraded executable: %v", err)
	}
	if string(data) != "new-binary" {
		t.Fatalf("unexpected upgraded binary contents: %q", string(data))
	}
}

func TestVersionLabel(t *testing.T) {
	tests := map[string]string{
		"":       "unknown",
		"dev":    "dev",
		"0.1.0":  "v0.1.0",
		"v0.1.0": "v0.1.0",
	}
	for input, want := range tests {
		if got := versionLabel(input); got != want {
			t.Fatalf("versionLabel(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	if got, ok := compareVersions("0.1.0", "0.2.0"); !ok || got >= 0 {
		t.Fatalf("expected 0.1.0 < 0.2.0, got (%d, %t)", got, ok)
	}
	if got, ok := compareVersions("0.2.0", "0.2.0"); !ok || got != 0 {
		t.Fatalf("expected equality, got (%d, %t)", got, ok)
	}
	if _, ok := compareVersions("dev", "0.2.0"); ok {
		t.Fatal("expected dev version comparison to be unsupported")
	}
}

func buildChecksums(assetName string, archive []byte) []byte {
	sum := sha256.Sum256(archive)
	return []byte(hex.EncodeToString(sum[:]) + "  " + assetName + "\n")
}

func buildArchive(t *testing.T, assetName, binaryName string, data []byte) []byte {
	t.Helper()
	if strings.HasSuffix(assetName, ".zip") {
		var buf bytes.Buffer
		writer := zip.NewWriter(&buf)
		file, err := writer.Create(binaryName)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := file.Write(data); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("close zip writer: %v", err)
		}
		return buf.Bytes()
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	header := &tar.Header{
		Name: binaryName,
		Mode: 0o755,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatalf("write tar file: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	return buf.Bytes()
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func bytesResponse(body []byte) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}
}
