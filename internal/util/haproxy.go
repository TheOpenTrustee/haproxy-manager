package util

import (
	"context"
	"fmt"
	"log"

	client_native "github.com/haproxytech/client-native/v6"
	"github.com/haproxytech/client-native/v6/configuration"
	cfg_opt "github.com/haproxytech/client-native/v6/configuration/options"
	"github.com/haproxytech/client-native/v6/models"
	"github.com/haproxytech/client-native/v6/options"
	runtime_api "github.com/haproxytech/client-native/v6/runtime"
	runtime_options "github.com/haproxytech/client-native/v6/runtime/options"
)

func CreateClient() (client_native.HAProxyClient, error) {
	haproxyOptions := cfg_opt.ConfigurationOptions{}
	haproxyOptions.ConfigurationFile = "/etc/haproxy/haproxy.cfg"
	haproxyOptions.Haproxy = "/usr/sbin/haproxy"
	haproxyOptions.BackupsNumber = 9
	haproxyOptions.TransactionDir = "/etc/haproxy/transactions"

	confClient, err := configuration.New(context.Background(),
		cfg_opt.ConfigurationFile(haproxyOptions.ConfigurationFile),
		cfg_opt.HAProxyBin(haproxyOptions.Haproxy),
		cfg_opt.Backups(haproxyOptions.BackupsNumber),
		cfg_opt.UsePersistentTransactions,
		cfg_opt.TransactionsDir(haproxyOptions.TransactionDir),
		cfg_opt.MasterWorker,
		cfg_opt.UseMd5Hash,
	)
	if err != nil {
		return nil, fmt.Errorf("error setting up configuration client: %s", err.Error())
	}

	ctx := context.Background()
	ms := runtime_options.MasterSocket("/var/run/haproxy.sock")
	runtimeClient, err := runtime_api.New(ctx, ms)
	if err != nil {
		return nil, fmt.Errorf("error setting up runtime client: %s", err.Error())
	}

	opt := []options.Option{
		options.Configuration(confClient),
		options.Runtime(runtimeClient),
	}

	client, err := client_native.New(ctx, opt...)
	if err != nil {
		log.Fatalf("Error initializing configuration client: %v", err)
	}

	return client, nil
}

func GetFrontends() models.Frontends {
	c, err := CreateClient()
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	configuration, err := c.Configuration()
	if err != nil {
		log.Fatalf("Error getting configuration: %v", err)
	}

	ru, err := c.Runtime()
	if err != nil {
		log.Fatalf("Error getting runtime: %v", err)
	}

	i, err := configuration.GetVersion("")
	if err != nil {
		log.Fatalf("Error getting version: %v", err)
	}

	tid, err := configuration.StartTransaction(i)
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}

	// We don't need a transaction id unless we're making changes to the configuration
	// tid := &models.Transaction{
	// 	ID: "",
	// }

	processInfo, err := ru.GetInfo()
	if err != nil {
		log.Fatalf("Error getting process info: %v", err)
	}

	_, backends, err := configuration.GetBackends(tid.ID)
	if err != nil {
		log.Fatalf("Error getting backends: %v", err)
	}

	_, frontends, err := configuration.GetFrontends(tid.ID)
	if err != nil {
		log.Fatalf("Error getting frontends: %v", err)
	}

	// err = configuration.CreateFrontend(&models., tid.ID, 1)
	// if err != nil {
	// 	log.Fatalf("Error creating frontend: %v", err)
	// }

	_, _ = configuration.CommitTransaction(tid.ID)

	fmt.Printf("%v\n", processInfo.Info)
	fmt.Printf("%v\n", ru.GetStats())
	fmt.Printf("%v\n", backends[0].Name)
	fmt.Printf("%v\n", frontends)

	return frontends

}
