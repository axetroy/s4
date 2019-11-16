package command

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"syscall"

	"github.com/cheggaaa/pb/v3"
	"github.com/fatih/color"
)

func Upgrade() error {

	res, err := http.Get("https://api.github.com/repos/axetroy/s4/releases/latest")

	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return err
	}

	type Asset struct {
		Name               string `json:"name"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	}

	type Response struct {
		TagName string  `json:"tag_name"`
		Assets  []Asset `json:"assets"`
	}

	response := Response{}

	if err = json.Unmarshal(body, &response); err != nil {
		return err
	}

	executablePath, err := os.Executable()

	if err != nil {
		return err
	}

	var currentAsset *Asset

	filename := fmt.Sprintf("s4_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)

	for _, asset := range response.Assets {
		if asset.Name == filename {
			currentAsset = &asset
			break
		}
	}

	if currentAsset == nil {
		fmt.Println("Can not found version for your platform")
		return nil
	}

	cmd := exec.Command(executablePath, "version")

	cmdOutput, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	currentVersion := "v" + string(cmdOutput)

	if response.TagName == currentVersion {
		fmt.Printf("You are using the latest version `%s`\n", color.GreenString(response.TagName))
		//return nil
	}

	fmt.Printf("Upgrading from `%s` to `%s` ...\n", color.GreenString(currentVersion), color.YellowString(response.TagName))

	// create temp dir to system
	tempDir, err := ioutil.TempDir("", "s4_download_")

	if err != nil {
		return err
	}

	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, os.Kill, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-quit
		_ = os.RemoveAll(tempDir)
		os.Exit(1)
	}()

	defer os.RemoveAll(tempDir)

	fileName := path.Join(tempDir, response.TagName+"_"+currentAsset.Name)

	if err := DownloadFile(fileName, currentAsset.BrowserDownloadUrl); err != nil {
		return err
	}

	// decompress the tag
	if err := decompress(fileName, path.Join(tempDir, response.TagName+"_")); err != nil {
		return err
	}

	destFilename := path.Join(tempDir, response.TagName+"_s4")

	// cover the binary file
	if err := os.Rename(destFilename, executablePath); err != nil {
		return err
	}

	ps := exec.Command(executablePath, "--help")

	ps.Stderr = os.Stderr
	ps.Stdout = os.Stdout

	if err := ps.Run(); err != nil {
		return err
	}

	return nil
}

// decompress gzip
func decompress(tarFile, dest string) error {
	srcFile, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		filename := dest + hdr.Name

		file, err := os.Create(filename)

		if err != nil {
			return err
		}

		if err := file.Chmod(os.FileMode(hdr.Mode)); err != nil {
			return err
		}

		if _, err := io.Copy(file, tr); err != nil {
			return err
		}
	}
	return nil
}

func DownloadFile(filepath string, url string) error {
	tmpl := fmt.Sprintf(`{{string . "prefix"}}{{ green "%s" }} {{counters . }} {{ bar . "[" "=" ">" "-" "]"}} {{percent . }} {{speed . }}{{string . "suffix"}}`, filepath)

	// Get the data
	response, err := http.Get(url)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	// Create the file
	writer, err := os.Create(filepath)

	if err != nil {
		return err
	}

	defer writer.Close()

	bar := pb.ProgressBarTemplate(tmpl).Start64(response.ContentLength)

	bar.SetWriter(os.Stdout)

	barReader := bar.NewProxyReader(response.Body)

	_, err = io.Copy(writer, barReader)

	bar.Finish()

	return err
}
