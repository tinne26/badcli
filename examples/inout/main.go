package main

import "fmt"
import "github.com/tinne26/badcli"

func main() {
	cli := badcli.NewCLI("inout", "Given some flags, inout prints the passed values.")
	cli.AddUsageSection("Additional usage section. Nothing really interesting to say.")
	cli.RegisterFlag("color" , "Color in hex or rgb format.", badcli.NewColorString(0, 0, 0), 'c')
	cli.RegisterFlag("number", "Number between 11 and 99.", badcli.NewBoundedInt(0, 11, 99), 'n')
	//cli.RegisterFlag("regexp" , "Any string ~= /[a-zA-Z0-9]{1-9}/.", badcli.NewRegexp(`[a-zA-Z0-9]{1-9}`))
	cli.ParseArguments()

	// show the values of each flag
	err := cli.EachFlag(func(fullFlagName string, value badcli.FlagValue) error {
		if cli.FlagSetByUser(fullFlagName) {
			fmt.Printf("Value of '--%s': %s\n", fullFlagName, value)
		} else {
			fmt.Printf("Value of '--%s': unset\n", fullFlagName)
		}
		return nil
	})
	if err != nil {
		panic("can't happen unless the function passed to cli.EachFlagName() fails")
	}

	// if cli.FlagSetByUser("number") {
	// 	fmt.Printf("Value of '--number': %s\n", cli.GetFlagValue("color").(*ColorString).String())
	// } else {
	// 	fmt.Printf("Value of '--number': unset\n")
	// }
}
