// Command build provides the repository's shell-independent development tasks.
package main

import (
	"archive/zip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

var semanticTag = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

func main() {
	if err := runTask(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "build:", err)
		os.Exit(1)
	}
}

func runTask(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	root, err := findRoot()
	if err != nil {
		return err
	}

	switch args[0] {
	case "check":
		if len(args) != 1 {
			return usageError()
		}
		return check(root)
	case "actions":
		if len(args) != 1 {
			return usageError()
		}
		return lintActions(root)
	case "build":
		if len(args) > 2 {
			return usageError()
		}
		if err := check(root); err != nil {
			return err
		}
		version := ""
		if len(args) == 2 {
			version = args[1]
		}
		path, err := buildWindows(root, filepath.Join(root, "dist", "odhydrate.exe"), version)
		if err != nil {
			return err
		}
		hash, err := fileSHA256(path)
		if err != nil {
			return err
		}
		fmt.Printf("SHA256  %X  %s\n", hash, path)
		return nil
	case "release":
		tag, err := releaseTag(args)
		if err != nil {
			return err
		}
		_, err = createReleaseAssets(root, tag)
		return err
	case "publish":
		tag, err := releaseTag(args)
		if err != nil {
			return err
		}
		return publishRelease(root, tag)
	case "help", "-h", "--help":
		fmt.Print(usage())
		return nil
	default:
		return usageError()
	}
}

func releaseTag(args []string) (string, error) {
	if len(args) == 2 {
		return args[1], nil
	}
	if len(args) == 1 {
		if tag := os.Getenv("TAG"); tag != "" {
			return tag, nil
		}
	}
	return "", usageError()
}

func usageError() error {
	return errors.New("invalid arguments\n\n" + usage())
}

func usage() string {
	return `Usage: go run ./tools/build <task> [argument]

Tasks:
  check                Check formatting, run tests, and vet the Windows target
  actions              Lint GitHub Actions workflows with actionlint
  build [version]      Check and build dist/odhydrate.exe
  release <tag>        Check and create a versioned ZIP and SHA256SUMS.txt
  publish <tag>        Publish already-created release assets with gh
`
}

func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("could not find go.mod")
		}
		dir = parent
	}
}

func check(root string) error {
	fmt.Println("==> Check formatting")
	cmd := exec.Command("gofmt", "-l", ".")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gofmt: %w\n%s", err, output)
	}
	if unformatted := strings.TrimSpace(string(output)); unformatted != "" {
		return fmt.Errorf("gofmt required for:\n%s", unformatted)
	}

	fmt.Println("==> Test")
	if err := run(root, nil, "go", "test", "./..."); err != nil {
		return err
	}

	fmt.Println("==> Vet Windows amd64")
	return run(root, map[string]string{"GOOS": "windows", "GOARCH": "amd64"}, "go", "vet", "./...")
}

func lintActions(root string) error {
	files, err := workflowFiles(root)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("no GitHub Actions workflows found")
	}
	fmt.Println("==> Lint GitHub Actions")
	return run(root, nil, "actionlint", files...)
}

func workflowFiles(root string) ([]string, error) {
	dir := filepath.Join(root, ".github", "workflows")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".yml" || ext == ".yaml" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

func buildWindows(root, output, version string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return "", err
	}
	args := []string{"build", "-trimpath"}
	ldflags := "-s -w"
	if version != "" {
		ldflags += " -X main.version=" + version
	}
	args = append(args, "-ldflags", ldflags, "-o", output, "./src")
	fmt.Println("==> Build Windows amd64")
	if err := run(root, map[string]string{"GOOS": "windows", "GOARCH": "amd64"}, "go", args...); err != nil {
		return "", err
	}
	return output, nil
}

type releaseAssets struct {
	archive   string
	checksums string
}

func createReleaseAssets(root, tag string) (releaseAssets, error) {
	var assets releaseAssets
	version, err := validateTag(root, tag)
	if err != nil {
		return assets, err
	}
	if err := check(root); err != nil {
		return assets, err
	}

	dist := filepath.Join(root, "dist")
	if err := os.MkdirAll(dist, 0o755); err != nil {
		return assets, err
	}
	temp, err := os.MkdirTemp(dist, ".release-")
	if err != nil {
		return assets, err
	}
	defer os.RemoveAll(temp)

	exe, err := buildWindows(root, filepath.Join(temp, "odhydrate.exe"), version)
	if err != nil {
		return assets, err
	}

	base := "odhydrate-windows-amd64-" + version
	assets.archive = filepath.Join(dist, base+".zip")
	files := []archiveFile{
		{source: exe, name: filepath.ToSlash(filepath.Join(base, "odhydrate.exe"))},
		{source: filepath.Join(root, "LICENSE"), name: filepath.ToSlash(filepath.Join(base, "LICENSE"))},
		{source: filepath.Join(root, "README.md"), name: filepath.ToSlash(filepath.Join(base, "README.md"))},
		{source: filepath.Join(root, "README.zh-CN.md"), name: filepath.ToSlash(filepath.Join(base, "README.zh-CN.md"))},
	}
	if err := writeZIP(assets.archive, files); err != nil {
		return assets, err
	}

	hash, err := fileSHA256(assets.archive)
	if err != nil {
		return assets, err
	}
	assets.checksums = filepath.Join(dist, "SHA256SUMS.txt")
	line := fmt.Sprintf("%x  %s\n", hash, filepath.Base(assets.archive))
	if err := os.WriteFile(assets.checksums, []byte(line), 0o644); err != nil {
		return assets, err
	}
	fmt.Println("Created", assets.archive)
	fmt.Println("Created", assets.checksums)
	return assets, nil
}

func publishRelease(root, tag string) error {
	version, err := validateTag(root, tag)
	if err != nil {
		return err
	}
	dist := filepath.Join(root, "dist")
	archive := filepath.Join(dist, "odhydrate-windows-amd64-"+version+".zip")
	checksums := filepath.Join(dist, "SHA256SUMS.txt")
	for _, path := range []string{archive, checksums} {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("release asset %s: %w", path, err)
		}
	}
	target, err := commandOutput(root, "git", "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	return run(root, nil, "gh", "release", "create", tag, archive, checksums,
		"--target", target, "--generate-notes", "--title", "odhydrate "+tag)
}

func validateTag(_ string, tag string) (string, error) {
	if !semanticTag.MatchString(tag) {
		return "", fmt.Errorf("expected a semantic-version tag such as v0.1.0; got %q", tag)
	}
	return strings.TrimPrefix(tag, "v"), nil
}

type archiveFile struct {
	source string
	name   string
}

func writeZIP(path string, files []archiveFile) (err error) {
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := output.Close(); err == nil {
			err = closeErr
		}
	}()

	zw := zip.NewWriter(output)
	defer func() {
		if closeErr := zw.Close(); err == nil {
			err = closeErr
		}
	}()
	for _, file := range files {
		if err := addToZIP(zw, file); err != nil {
			return err
		}
	}
	return nil
}

func addToZIP(zw *zip.Writer, file archiveFile) error {
	info, err := os.Stat(file.source)
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = file.name
	header.Method = zip.Deflate
	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	r, err := os.Open(file.source)
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(w, r)
	return err
}

func fileSHA256(path string) ([sha256.Size]byte, error) {
	var result [sha256.Size]byte
	f, err := os.Open(path)
	if err != nil {
		return result, err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return result, err
	}
	copy(result[:], hash.Sum(nil))
	return result, nil
}

func run(root string, overrides map[string]string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = root
	cmd.Env = environment(overrides)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func commandOutput(root, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %w\n%s", name, err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

func environment(overrides map[string]string) []string {
	if len(overrides) == 0 {
		return os.Environ()
	}
	env := os.Environ()
	for key, value := range overrides {
		prefix := key + "="
		for i := len(env) - 1; i >= 0; i-- {
			candidate := env[i]
			match := strings.HasPrefix(candidate, prefix)
			if runtime.GOOS == "windows" {
				match = strings.EqualFold(strings.SplitN(candidate, "=", 2)[0], key)
			}
			if match {
				env = append(env[:i], env[i+1:]...)
			}
		}
		env = append(env, key+"="+value)
	}
	return env
}
