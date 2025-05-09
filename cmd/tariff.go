package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/evcc-io/evcc/api"
	"github.com/spf13/cobra"
)

// tariffCmd represents the vehicle command
var tariffCmd = &cobra.Command{
	Use:   "tariff [name]",
	Short: "Query configured tariff",
	Args:  cobra.MaximumNArgs(1),
	Run:   runTariff,
}

func init() {
	rootCmd.AddCommand(tariffCmd)
}

func runTariff(cmd *cobra.Command, args []string) {
	// load config
	if err := loadConfigFile(&conf, !cmd.Flag(flagIgnoreDatabase).Changed); err != nil {
		fatal(err)
	}

	// setup environment
	if err := configureEnvironment(cmd, &conf); err != nil {
		fatal(err)
	}

	tariffs, err := configureTariffs(&conf.Tariffs)
	if err != nil {
		fatal(err)
	}

	var name string
	if len(args) == 1 {
		name = args[0]
	}

	for u, tf := range map[api.TariffUsage]api.Tariff{
		api.TariffUsageGrid:    tariffs.Grid,
		api.TariffUsageFeedIn:  tariffs.FeedIn,
		api.TariffUsageCo2:     tariffs.Co2,
		api.TariffUsagePlanner: tariffs.Planner,
		api.TariffUsageSolar:   tariffs.Solar,
	} {
		key := u.String()
		if name != "" && key != name {
			continue
		}

		if tf == nil {
			continue
		}

		if name == "" {
			fmt.Println(key + ":")
		}

		rates, err := tf.Rates()
		if err != nil {
			fatal(err)
		}

		unit := "Price/Cost"
		switch tf.Type() {
		case api.TariffTypeCo2:
			unit += "Footprint (gCO2/kWh)"
		case api.TariffTypeSolar:
			unit = "Yield (W)"
		default:
			if c := conf.Tariffs.Currency; c != "" {
				unit += fmt.Sprintf(" (%s/kWh)", c)
			}
		}

		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		fmt.Fprintln(tw, "From\tTo\t"+unit)
		const format = "2006-01-02 15:04:05"

		for _, r := range rates {
			fmt.Fprintf(tw, "%s\t%s\t%.3f\n", r.Start.Local().Format(format), r.End.Local().Format(format), r.Value)
		}
		tw.Flush()

		fmt.Println()
	}

	// wait for shutdown
	<-shutdownDoneC()
}
