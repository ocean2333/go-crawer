package album

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ocean2333/go-crawer/src/model"
	"github.com/spf13/cobra"
)

func SearchPageCmd() *cobra.Command {
	var (
		rid      string
		keywords string
		page     uint32
	)
	cmd := &cobra.Command{
		Use:   "search",
		Short: "send a search page request to crawer",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:10320/album/search_page?rid=%s&keywords=%s&page=%d", rid, keywords, page), nil)
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
	flags.StringVar(&keywords, "keywords", "", "keywords")
	flags.Uint32Var(&page, "page", 1, "page")

	if cmd.MarkFlagRequired("rid") != nil {
		return nil
	}
	if cmd.MarkFlagRequired("keywords") != nil {
		return nil
	}

	return cmd
}
