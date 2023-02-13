package main

import (
	"fmt"
	"log"
	"os"
	"time"

	vault "github.com/hashicorp/vault/api"
)

func init() {
	initConfiguration()
}

func main() {
	err := exec()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}

	os.Exit(0)
}

func exec() error {
	cf, err := processConfiguration()

	if err != nil {
		return err
	}

	unsealKey, err := cf.getUnsealKey()

	if err != nil {
		return err
	}

	vaultConfig := vault.DefaultConfig()
	vaultConfig.Address = cf.Address

	vaultClient, err := vault.NewClient(vaultConfig)

	if err != nil {
		return fmt.Errorf("%s %s", "Unable to initialize Vault client: %v", err)
	}

	vaultClientSystem := vaultClient.Sys()

	timeSleep := time.Duration(cf.LoopSeconds) * time.Second

	for {
		time.Sleep(timeSleep)

		sealStatusResponse, err := vaultClientSystem.SealStatus()

		if err != nil {
			log.Printf("Unable to check Vault seal status: %v\n", err)

			continue
		}

		if !sealStatusResponse.Sealed {
			log.Println("Vault is already unsealed")

			continue
		}

		log.Println("Vault is sealed; attempting unseal")

		sealStatusResponse, err = vaultClientSystem.Unseal(unsealKey)

		if err != nil {
			log.Printf("Unable to unseal Vault: %v\n", err)

			continue
		}

		if sealStatusResponse.Sealed {
			log.Println("Unable to unseal Vault")

			continue
		}

		log.Println("Vault is unsealed; unseal successful")
	}
}
