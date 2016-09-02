package secrets

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/docker/swarmkit/api"
	"github.com/docker/swarmkit/cmd/swarmctl/common"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

type secretSorter []*api.Secret

func (k secretSorter) Len() int      { return len(k) }
func (k secretSorter) Swap(i, j int) { k[i], k[j] = k[j], k[i] }
func (k secretSorter) Less(i, j int) bool {
	iTime := time.Unix(k[i].Meta.CreatedAt.Seconds, int64(k[i].Meta.CreatedAt.Nanos))
	jTime := time.Unix(k[j].Meta.CreatedAt.Seconds, int64(k[j].Meta.CreatedAt.Nanos))
	return jTime.Before(iTime)
}

var (
	listCmd = &cobra.Command{
		Use:   "ls",
		Short: "List secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("ls command takes no arguments")
			}

			flags := cmd.Flags()
			quiet, err := flags.GetBool("quiet")
			if err != nil {
				return err
			}

			client, err := common.Dial(cmd)
			if err != nil {
				return err
			}

			resp, err := client.ListSecrets(common.Context(cmd), &api.ListSecretsRequest{})
			if err != nil {
				return err
			}

			var output func(*api.Secret)

			if !quiet {
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
				defer func() {
					// Ignore flushing errors - there's nothing we can do.
					_ = w.Flush()
				}()
				common.PrintHeader(w, "ID", "Name", "Created", "Digest", "Size")
				output = func(s *api.Secret) {
					created := time.Unix(int64(s.Meta.CreatedAt.Seconds), int64(s.Meta.CreatedAt.Nanos))
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n",
						s.ID,
						s.Spec.Annotations.Name,
						humanize.Time(created),
						s.Digest,
						s.SecretSize,
					)
				}

			} else {
				output = func(s *api.Secret) { fmt.Println(s.ID) }
			}

			sorted := secretSorter(resp.Secrets)
			sort.Sort(sorted)
			for _, s := range sorted {
				output(s)
			}
			return nil
		},
	}
)

func init() {
	listCmd.Flags().BoolP("quiet", "q", false, "Only display secret names")
}
