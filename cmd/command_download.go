package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const TMP_DOWNLOAD_DIR = "/tmp/downloads"

// Download the URL and write the file to the given destination, with owner and permissions set accordingly.
// Destination can either be an existing directory or a file. If a directory is given the downloaded file will moved
// there using the file name from the URL. If it is a file, the downloaded file will be moved (and possibly renamed) to
// that destination. If the extract flag is set the downloaded file will be extracted to the directory given in the
// destination field.
type DownloadCommand struct {
	Url         string      // Where to download from.
	Destination string      // Where to put the downloaded file.
	Owner       string      // Owner of the downloaded file.
	Permissions os.FileMode // Permissions of the downloaded file.
	Extract     bool        // Extract the downloaded archive.
}

// Download the file from the given URL and extract it to the given directory. If the directory does not exist it is
// created. See the "ExtractFile" command for a list of supported archive types.
func DownloadAndExtract(url, destination string) *DownloadCommand {
	if url == "" {
		panic("empty url given")
	}

	if destination == "" {
		panic("no destination given")
	}

	return &DownloadCommand{Url: url, Destination: destination, Extract: true}
}

func DownloadToFile(url, destination, owner string, permissions os.FileMode) *DownloadCommand {
	if url == "" {
		panic("empty url given")
	}

	if destination == "" {
		panic("no destination given")
	}

	return &DownloadCommand{Url: url, Destination: destination, Owner: owner, Permissions: permissions}
}

func (dc *DownloadCommand) Docker() string {
	return fmt.Sprintf("RUN %s", dc.Shell())
}

func (dc *DownloadCommand) Shell() string {
	if dc.Url == "" {
		panic("empty url given")
	}

	filename := path.Base(dc.Url)
	destination := fmt.Sprintf("%s/%s", TMP_DOWNLOAD_DIR, filename)

	cmd := []string{}

	cmd = append(cmd, fmt.Sprintf("mkdir -p %s", TMP_DOWNLOAD_DIR))
	cmd = append(cmd, fmt.Sprintf("cd %s", TMP_DOWNLOAD_DIR))
	cmd = append(cmd, fmt.Sprintf(`curl -SsfLO "%s"`, dc.Url))

	switch {
	case dc.Extract && dc.Destination == "":
		panic(fmt.Errorf("shall extract, but don't know where (i.e. destination field is empty"))
	case dc.Extract:
		cmd = append(cmd, ExtractFile(destination, dc.Destination).Shell())
	case dc.Destination != "":
		cmd = append(cmd, fmt.Sprintf("mv %s %s", destination, dc.Destination))
		destination = dc.Destination
	}

	if dc.Owner != "" && dc.Owner != "root" {
		ifFile := fmt.Sprintf("{ [[ -f %s ]] && chown %s %s; }", destination, dc.Owner, destination)
		ifInDir := fmt.Sprintf("{ [[ -d %s ]] && [[ -f %s/%s ]] && chown %s %s/%s; }", destination, destination, filename, dc.Owner, destination, filename)
		ifDir := fmt.Sprintf("{ [[ -d %s ]] && chown -R %s %s; }", destination, dc.Owner, destination)
		err := `{ echo "Couldn't determine target" && exit 1; }`
		cmd = append(cmd, fmt.Sprintf("{ %s; }", strings.Join([]string{ifFile, ifInDir, ifDir, err}, " || ")))
	}

	if dc.Permissions != 0 {
		ifFile := fmt.Sprintf("{ [[ -f %s ]] && chmod %o %s; }", destination, dc.Permissions, destination)
		ifInDir := fmt.Sprintf("{ [[ -d %s ]] && [[ -f %s/%s ]] && chmod %o %s/%s; }", destination, destination,
			filename, dc.Permissions, destination, filename)
		ifDir := fmt.Sprintf("{ [[ -d %s ]] && chmod %o %s; }", destination, dc.Permissions, destination)
		err := `{ echo "Couldn't determine target" && exit 1; }`
		cmd = append(cmd, fmt.Sprintf("{ %s; }", strings.Join([]string{ifFile, ifInDir, ifDir, err}, " || ")))
	}

	return strings.Join(cmd, " && ")
}

func (dc *DownloadCommand) Logging() string {
	sList := []string{"[DWNLOAD]"}

	if dc.Owner != "" && dc.Owner != "root" {
		sList = append(sList, fmt.Sprintf("[CHOWN:%s]", dc.Owner))
	}

	if dc.Permissions != 0 {
		sList = append(sList, fmt.Sprintf("[CHMOD:%.4o]", dc.Permissions))
	}

	sList = append(sList, fmt.Sprintf(" >> downloading file %q", dc.Url))
	if dc.Extract {
		sList = append(sList, " and extracting archive")
	}
	if dc.Destination != "" {
		sList = append(sList, fmt.Sprintf(" to %q", dc.Destination))
	}
	return strings.Join(sList, "")
}