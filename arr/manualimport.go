package arr

import "context"

// manualImportCommand is the body of the ManualImport command the *arr services accept.
type manualImportCommand[M ~string, F any] struct {
	Name       string `json:"name"`
	ImportMode M      `json:"importMode"`
	Files      []F    `json:"files"`
}

// SendManualImport posts a ManualImport command for files to the command endpoint at path.
// It returns errNoFiles when files is empty and errMode when mode is not auto, move or copy,
// so each service keeps its own sentinel errors.
func SendManualImport[M ~string, F any](ctx context.Context, c *BaseClient, path string, mode M, files []F, errNoFiles, errMode error) (*CommandResponse, error) {
	if len(files) == 0 {
		return nil, errNoFiles
	}
	switch string(mode) {
	case "auto", "move", "copy":
	default:
		return nil, errMode
	}
	body := manualImportCommand[M, F]{Name: "ManualImport", ImportMode: mode, Files: files}
	return PostOne[CommandResponse](ctx, c, path, body)
}
