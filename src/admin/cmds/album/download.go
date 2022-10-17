package album

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ocean2333/go-crawer/src/model"
	"github.com/spf13/cobra"
)

func DownloadCmd() *cobra.Command {
	var (
		rid string
		aid string
	)
	cmd := &cobra.Command{
		Use:   "download",
		Short: "send a download page request to crawer",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:10320/album/download?rid=%s&aid=%s", rid, aid), nil)
			if err != nil {
				return err
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("status code: %d", resp.StatusCode)
			}
			data, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return err
			}
			titles := new(model.AdminTitlesResponse)
			json.Unmarshal(data, titles)
			fmt.Println(titles)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&rid, "rid", "", "rid")
	flags.StringVar(&aid, "aid", "", "aid")

	if cmd.MarkFlagRequired("rid") != nil {
		return nil
	}
	if cmd.MarkFlagRequired("aid") != nil {
		return nil
	}

	return cmd
}
