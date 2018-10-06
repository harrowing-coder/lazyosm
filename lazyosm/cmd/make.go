package cmd

import (
	"fmt"
	"github.com/murphy214/lazyosm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var tile string
var resolution int

func init() {
	rootCmd.AddCommand(drawCmd)
	rootCmd.PersistentFlags().StringVarP(&tile, "outfilename", "o", "out.geobuf", "outfilename")
	viper.BindPFlag("outfilename", rootCmd.PersistentFlags().Lookup("outfilename"))
	rootCmd.PersistentFlags().IntVarP(&resolution, "limit", "l", 1000, "limit of blocks to be open")
	viper.BindPFlag("limit", rootCmd.PersistentFlags().Lookup("limit"))
}

var drawCmd = &cobra.Command{
	Use:   "make",
	Short: "Creates an output geobuf file for the given osm data.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.HasSuffix(tile, "geojson") {
			tmpfile, _ := ioutil.TempFile("", "example")

			defer os.Remove(tmpfile.Name()) // clean up
			osm.MakeOutputGeobuf(filename, tmpfile.Name(), resolution)
			val, err := exec.Command("geobuf2geojson", tmpfile.Name(), tile).Output()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(val))
		} else {
			osm.MakeOutputGeobuf(filename, tile, resolution)
		}
	},
}
