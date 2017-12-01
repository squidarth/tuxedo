package main

import (
	"fmt"
	"io/ioutil"
	"os"
	flag "github.com/ogier/pflag"
	"golang.org/x/crypto/ssh"
	"log"
  "github.com/hashicorp/hcl"
)

var (
	dryRun bool
)

func parseTux(content string) (*Config, error) {
  config := &Config{}

  hclParseTree, err := hcl.Parse(content)
  if err != nil {
    return nil, err
  }
	if err := hcl.DecodeObject(&config, hclParseTree); err != nil {
		return nil, err
	}
  return config, nil
}

func main() {
	flag.Parse()
	if flag.NArg() > 0 {
		flag.Usage()
		os.Exit(1)
	}


  tuxContent, err := ioutil.ReadFile("jenkins.tux")
	if err != nil {
		fmt.Println("Failed to find a jenkins.tux file.")
	}

  config, err := parseTux(string(tuxContent))
  if err != nil {
		fmt.Println("Theres an error in the tux config.")
  }

	privateKey, err := ioutil.ReadFile(config.SSHSettingsConfig.PathToKey)
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		fmt.Println("Failed to parse the private key file")
		fmt.Println(err.Error())
	}

  client, err := ssh.Dial("tcp", config.SSHSettingsConfig.HostIp + ":22", &ssh.ClientConfig{
		User: config.SSHSettingsConfig.SSHUsername,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // FIXME: please be more secure in checking host keys
	})
	if err != nil {
		log.Fatalln("Failed to dial:", err)
	}

  if (dryRun) {
    generalDiff, err := handleGeneralSettingsDryRun(client, &config.GeneralConfig)
    if err != nil {
      fmt.Println("Failed parsing general settings")
      return
    }

    securityDiff, err := handleSecuritySettingsDryRun(client, &config.SecurityConfig)
    if err != nil {
      fmt.Println("Failed parsing security settings")
      return
    }

    diff := generalDiff + securityDiff
    if diff == "" {
      fmt.Println("No changes.")
    } else {
      fmt.Println("Tuxedo Changes:\n\n" + diff)
    }
  } else {
    err := handleSecuritySettings(client, &config.SecurityConfig)
    if err != nil {
      fmt.Println("Error Applying Security Changes")
      return
    }

    general_err := handleGeneralSettings(client, &config.GeneralConfig)
    if general_err != nil {
      fmt.Println("Error Applying General Changes")
      return
    }

    fmt.Println("Changes Applied")
  }
}

func init() {
	// We pass the user variable we declared at the package level (above).
	// The "&" character means we are passing the variable "by reference" (as opposed to "by value"),
	// meaning: we don't want to pass a copy of the user variable. We want to pass the original variable.
	flag.BoolVarP(&dryRun, "dry-run", "d", false, "Run tuxedo in dry-run mode")
}
