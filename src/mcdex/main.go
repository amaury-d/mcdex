package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type command struct {
	Fn    func() error
	Usage string
}

var gCommands = map[string]command{
	"createPack": command{
		Fn:    cmdCreatePack,
		Usage: "Create a new mod pack",
	},
	"installPack": command{
		Fn:    cmdInstallPack,
		Usage: "Install a mod pack",
	},
	"installLocalPack": command{
		Fn:    cmdInstallLocalPack,
		Usage: "Install specified directory as a pack",
	},
	"update": command{
		Fn:    cmdUpdate,
		Usage: "Download latest index",
	},
	"info": command{
		Fn:    cmdInfo,
		Usage: "Show runtime info",
	},
	"registerMod": command{
		Fn:    cmdRegisterMod,
		Usage: "Register a curseforge mod with an existing pack",
	},
	"installMods": command{
		Fn:    cmdInstallMods,
		Usage: "Install all mods using the manifest",
	},
	"runServer": command{
		Fn:    cmdRunServer,
		Usage: "Run a minecraft server with an existing pack",
	},
}

func cmdCreatePack() error {
	if flag.NArg() < 4 {
		return fmt.Errorf("Insufficient arguments")
	}

	// Create a new pack directory
	cp, err := NewModPack(flag.Arg(1), "")
	if err != nil {
		return err
	}

	// Create the manifest for this new pack
	err = cp.createManifest(flag.Arg(1), flag.Arg(2), flag.Arg(3))
	if err != nil {
		return err
	}

	// Create the launcher profile (and install forge if necessary)
	err = cp.createLauncherProfile()
	if err != nil {
		return err
	}

	return nil
}

func cmdInstallPack() error {
	// If there are not enough arguments, bail
	if flag.NArg() < 3 {
		return fmt.Errorf("Insufficient arguments")
	}

	// Get ZIP file
	cp, err := NewModPack(flag.Arg(1), flag.Arg(2))
	if err != nil {
		return err
	}

	// Download the pack
	err = cp.download()
	if err != nil {
		return err
	}

	// Process manifest
	err = cp.processManifest()
	if err != nil {
		return err
	}

	// Create launcher profile
	err = cp.createLauncherProfile()
	if err != nil {
		return err
	}

	// Install plugins
	err = cp.installMods()
	if err != nil {
		return err
	}

	// Install overrides
	err = cp.installOverrides()
	if err != nil {
		return err
	}

	return nil
}

func cmdInstallLocalPack() error {
	// If there are not enough arguments, bail
	if flag.NArg() < 2 {
		return fmt.Errorf("Insufficient arguments")
	}

	name := flag.Arg(1)
	if name == "." {
		name, _ = os.Getwd()
	}
	name, _ = filepath.Abs(name)

	// Create the mod pack directory (if it doesn't already exist)
	cp, err := OpenModPack(name)
	if err != nil {
		return err
	}

	// Setup a launcher profile
	err = cp.createLauncherProfile()
	if err != nil {
		return err
	}

	// Install all the mods
	err = cp.installMods()
	if err != nil {
		return err
	}

	return nil
}

func cmdUpdate() error {
	db, err := NewDatabase()
	if err != nil {
		log.Fatalf("%+v\n", err)
	}

	return db.Download()
}

func cmdInfo() error {
	fmt.Printf("Env: %+v\n", env())
	return nil
}

func cmdInstallMods() error {
	if flag.NArg() < 1 {
		return fmt.Errorf("Insufficient arguments")
	}

	cp, err := OpenModPack(flag.Arg(1))
	if err != nil {
		return err
	}

	err = cp.installMods()
	if err != nil {
		return err
	}

	return nil
}

func cmdRegisterMod() error {
	if flag.NArg() < 3 {
		return fmt.Errorf("Insufficient arguments")
	}

	if !strings.Contains(flag.Arg(2), "minecraft.curseforge.com") && flag.NArg() < 4 {
		return fmt.Errorf("Insufficient arguments")
	}

	cp, err := OpenModPack(flag.Arg(1))
	if err != nil {
		return err
	}

	err = cp.registerMod(flag.Arg(2), flag.Arg(3))
	if err != nil {
		return err
	}

	return nil
}

func cmdRunServer() error {
	if flag.NArg() < 2 {
		return fmt.Errorf("Insufficient arguments")
	}

	// Open the pack
	cp, err := OpenModPack(flag.Arg(1))
	if err != nil {
		return err
	}

	// Install the server jar, forge and dependencies
	err = cp.installServer()
	if err != nil {
		return err
	}

	return nil
	// Setup the command-line
	// java -jar <forge.jar>
}

func console(f string, args ...interface{}) {
	fmt.Printf(f, args...)
}

func usage() {
	console("usage: mcdex [<options>] <command> [<args>]\n")
	// console(" options:\n")
	// flag.PrintDefaults()
	console(" commands:\n")
	for id, cmd := range gCommands {
		console(" - %s: %s\n", id, cmd.Usage)
	}
}

func main() {
	// Process command-line args
	flag.Parse()
	if !flag.Parsed() || flag.NArg() < 1 {
		usage()
		os.Exit(-1)
	}

	// Initialize our environment
	err := initEnv()
	if err != nil {
		log.Fatalf("Failed to initialize: %s\n", err)
	}

	command, exists := gCommands[flag.Arg(0)]
	if !exists {
		console("ERROR: unknown command '%s'\n", flag.Arg(0))
		usage()
		os.Exit(-1)
	}

	err = command.Fn()
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
}

//mcdex update - download latest mcdex.sqlite
//mcdex forge.install <name> [<vsn>]
//mcdex forge.list

//mcdex init <name> <vsn> <desc>
//mcdex install <modname> [<vsn>]
