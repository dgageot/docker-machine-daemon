package handlers

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"strconv"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/commands"
)

// Create creates a Docker Machine
func Create(api libmachine.API, args map[string]string, form map[string][]string) (interface{}, error) {
	name, present := args["name"]
	if !present {
		return nil, errRequireMachineName
	}

	drivers, present := form["driver"]
	if !present || len(drivers) != 1 {
		return nil, errRequireDriverName
	}

	if err := createMachine(api, name, drivers[0], form); err != nil {
		return nil, err
	}

	return Success{"created", name}, nil
}

func createMachine(api libmachine.API, name string, driver string, form map[string][]string) error {
	validName := host.ValidateHostName(name)
	if !validName {
		return fmt.Errorf("Error creating machine: %s", mcnerror.ErrInvalidHostname)
	}

	exists, err := api.Exists(name)
	if err != nil {
		return fmt.Errorf("Error checking if host exists: %s", err)
	}
	if exists {
		return mcnerror.ErrHostAlreadyExists{
			Name: name,
		}
	}

	rawDriver, err := json.Marshal(&drivers.BaseDriver{
		MachineName: name,
		StorePath:   mcndirs.GetBaseDir(),
	})
	if err != nil {
		return fmt.Errorf("Error attempting to marshal bare driver data: %s", err)
	}

	h, err := api.NewHost(driver, rawDriver)
	if err != nil {
		return err
	}

	globalOpts := globalFlags{
		flags: form,
	}

	h.HostOptions = &host.Options{
		AuthOptions: &auth.Options{
			CertDir:          mcndirs.GetMachineCertDir(),
			CaCertPath:       filepath.Join(mcndirs.GetMachineCertDir(), "ca.pem"),
			CaPrivateKeyPath: filepath.Join(mcndirs.GetMachineCertDir(), "ca-key.pem"),
			ClientCertPath:   filepath.Join(mcndirs.GetMachineCertDir(), "cert.pem"),
			ClientKeyPath:    filepath.Join(mcndirs.GetMachineCertDir(), "key.pem"),
			ServerCertPath:   filepath.Join(mcndirs.GetMachineDir(), name, "server.pem"),
			ServerKeyPath:    filepath.Join(mcndirs.GetMachineDir(), name, "server-key.pem"),
			StorePath:        filepath.Join(mcndirs.GetMachineDir(), name),
			ServerCertSANs:   globalOpts.StringSlice("tls-san"),
		},
		EngineOptions: &engine.Options{
			ArbitraryFlags:   globalOpts.StringSlice("engine-opt"),
			Env:              globalOpts.StringSlice("engine-env"),
			InsecureRegistry: globalOpts.StringSlice("engine-insecure-registry"),
			Labels:           globalOpts.StringSlice("engine-label"),
			RegistryMirror:   globalOpts.StringSlice("engine-registry-mirror"),
			StorageDriver:    globalOpts.String("engine-storage-driver"),
			TLSVerify:        true,
			InstallURL:       globalOpts.String("engine-install-url"),
		},
		SwarmOptions: &swarm.Options{
			IsSwarm:        globalOpts.Bool("swarm"),
			Image:          globalOpts.String("swarm-image"),
			Master:         globalOpts.Bool("swarm-master"),
			Discovery:      globalOpts.String("swarm-discovery"),
			Address:        globalOpts.String("swarm-addr"),
			Host:           globalOpts.String("swarm-host"),
			Strategy:       globalOpts.String("swarm-strategy"),
			ArbitraryFlags: globalOpts.StringSlice("swarm-opt"),
			IsExperimental: globalOpts.Bool("swarm-experimental"),
		},
	}

	mcnFlags := h.Driver.GetCreateFlags()
	opts, err := parseFlags(form, mcnFlags, commands.SharedCreateFlags)
	if err != nil {
		return err
	}

	if err := h.Driver.SetConfigFromFlags(opts); err != nil {
		return fmt.Errorf("Error setting machine configuration from flags provided: %s", err)
	}

	if err := api.Create(h); err != nil {
		return err
	}

	if err := api.Save(h); err != nil {
		return fmt.Errorf("Error attempting to save store: %s", err)
	}

	return nil
}

func parseFlags(form map[string][]string, mcnflags []mcnflag.Flag, cliFlags []cli.Flag) (drivers.DriverOptions, error) {
	driverOpts := rpcdriver.RPCFlags{
		Values: make(map[string]interface{}),
	}

	for _, f := range cliFlags {
		switch f := f.(type) {
		case cli.StringFlag:
			driverOpts.Values[f.Name] = f.Value

			values, present := form[f.Name]
			if present && len(values) == 1 {
				driverOpts.Values[f.Name] = values[0]
			}
		case cli.StringSliceFlag:
			driverOpts.Values[f.Name] = f.Value.Value()

			values, present := form[f.Name]
			if present {
				driverOpts.Values[f.Name] = values
			}
		case cli.IntFlag:
			driverOpts.Values[f.Name] = f.Value

			values, present := form[f.Name]
			if present && len(values) == 1 {
				i, err := strconv.Atoi(values[0])
				if err != nil {
					driverOpts.Values[f.Name] = i
				}
			}
		case cli.BoolFlag:
			driverOpts.Values[f.Name] = false

			value, present := form[f.Name]
			if present && len(value) == 1 {
				driverOpts.Values[f.Name] = value[0]

				values, present := form[f.Name]
				if present && len(values) == 1 {
					driverOpts.Values[f.Name] = values[0] == "true"
				}
			}
		}
	}

	for _, f := range mcnflags {
		driverOpts.Values[f.String()] = f.Default()
		// Hardcoded logic for boolean... :(
		if f.Default() == nil {
			driverOpts.Values[f.String()] = false
		}

		values, present := form[f.String()]
		if present {
			switch f.(type) {
			case mcnflag.StringFlag:
				driverOpts.Values[f.String()] = values[0]
			case mcnflag.StringSliceFlag:
				driverOpts.Values[f.String()] = values
			case mcnflag.IntFlag:
				i, err := strconv.Atoi(values[0])
				if err != nil {
					return nil, err
				}

				driverOpts.Values[f.String()] = i
			case mcnflag.BoolFlag:
				driverOpts.Values[f.String()] = (values[0] == "true")
			}
		}
	}

	return driverOpts, nil
}

type globalFlags struct {
	flags map[string][]string
}

func (o *globalFlags) String(key string) string {
	value, present := o.flags[key]
	if present && len(value) > 0 {
		return value[0]
	}

	return ""
}

func (o *globalFlags) StringSlice(key string) []string {
	values, present := o.flags[key]
	if present {
		return values
	}

	return nil
}

func (o *globalFlags) Int(key string) int {
	value, present := o.flags[key]
	if present && len(value) > 0 {
		i, err := strconv.Atoi(value[0])
		if err != nil {
			return i
		}
	}

	return 0
}

func (o *globalFlags) Bool(key string) bool {
	value, present := o.flags[key]
	if present && len(value) > 0 {
		return value[0] == "true"
	}

	return false
}
