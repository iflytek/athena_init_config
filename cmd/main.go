package cmd

import (
	_ "embed"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xfyun/athena_init_config/jobs"
	"github.com/xfyun/athena_init_config/utils"
	"os"
)

//go:embed init.pub
var PublicKey string

var (
	// Used for flags.
	pub        string
	configAddr string
	userName   string
	password   string
	appId      string
	useHttps   bool
	rootCmd    = &cobra.Command{
		Use:   "init",
		Short: "Init job for athena serving config center",
		Long:  `Init job for athena serving config center `,
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
			var configUrl string
			if useHttps {
				configUrl = fmt.Sprintf("%s%s", "https://", configAddr)
			} else {
				configUrl = fmt.Sprintf("%s%s", "http://", configAddr)
			}
			configService, err := utils.NewCenterService(
				configUrl,
				appId,
				userName,
				password, pub)
			if err != nil {
				fmt.Println(err)
				return
			}
			err = jobs.Execute(configService)
			if err != nil {
				fmt.Println(err)
				return
			}
			return
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&userName, "username", "admin", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "config center api password")
	rootCmd.PersistentFlags().StringVar(&configAddr, "configAddr", "athena-polaris-cynosure:8099", "config center api url")
	rootCmd.PersistentFlags().BoolVar(&useHttps, "useHttps", false, "config http proto  center api url")

	rootCmd.PersistentFlags().StringVar(&appId, "appId", "123480995", "appId")

	rootCmd.PersistentFlags().StringVarP(&pub, "pub", "p", PublicKey, "name of license for the project")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")

	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("configAddr", rootCmd.PersistentFlags().Lookup("configAddr"))

	viper.SetDefault("username", "admin")
	rootCmd.MarkPersistentFlagRequired("password")

}
