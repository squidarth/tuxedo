package main

import (
	"bytes"
	"fmt"
	"github.com/clbanning/mxj"
	"io"
	"io/ioutil"
	"os"
	"path"
  "strconv"
//	"path/filepath"
	//	"github.com/hashicorp/hcl"
	//	"github.com/davecgh/go-spew/spew"
	shellquote "github.com/kballard/go-shellquote"
	flag "github.com/ogier/pflag"
	"golang.org/x/crypto/ssh"
	"log"
)

var (
	dryRun bool
)

func copyPath(filePath, destinationPath string, session *ssh.Session) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		return err
	}
	return copy(s.Size(), s.Mode().Perm(), path.Base(filePath), f, destinationPath, session)
}

func copy(size int64, mode os.FileMode, fileName string, contents io.Reader, destination string, session *ssh.Session) error {
	defer session.Close()
	w, err := session.StdinPipe()

	if err != nil {
		return err
	}

	cmd := shellquote.Join("sudo", "scp", "-t", destination)
	if err := session.Start(cmd); err != nil {
		w.Close()
		fmt.Println(err.Error())
		return err
	}

	errors := make(chan error)

	go func() {
		errors <- session.Wait()
	}()

	fmt.Fprintf(w, "C%#o %d %s\n", mode, size, fileName)
	_, copyErr := io.Copy(w, contents)
	if copyErr != nil {
		fmt.Println(copyErr.Error())
	}
	fmt.Fprint(w, "\x00")
	w.Close()

	return <-errors
}

func runShellCommand(cmd string, client *ssh.Client) string {
	session, err := client.NewSession()

	if err != nil {
		log.Fatalln("Failed to create session: " + err.Error())
	}
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(cmd); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	return b.String()
}

func readFile(fileName string, client *ssh.Client) string {
	cmd := shellquote.Join("cat", fileName)
	return runShellCommand(cmd, client)
}

func redString(str string) string {
  return "\u001b[31m" + str + "\u001b[0m"
}

func greenString(str string) string {
  return "\u001b[32m" + str + "\u001b[0m"
}

func handleSecuritySettingsDryRun(client *ssh.Client, securitySettings *Security) (string, error) {
	currentConfigStr := readFile("/var/lib/jenkins/config.xml", client)
	mv, _ := mxj.NewMapXml([]byte(currentConfigStr))

  diff := ""

  disableSignupValues, err := mv.ValuesForKey("disableSignup")
  if err != nil {
    fmt.Println("Couldn't get disableSignup")
    return "", err
  }

  disableSignup := disableSignupValues[0].(string)
  if strconv.FormatBool(securitySettings.DisableSignup) != disableSignup {
    diff += greenString("+ disable_signup " + strconv.FormatBool(securitySettings.DisableSignup)) + redString("\n- disable_signup " + disableSignup) + "\n"
  }

  disableRememberMeValues, err := mv.ValuesForKey("disableRememberMe")
  if err != nil {
    fmt.Println("Couldn't get disableSignup")
    return "", err
  }

  disableRememberMe := disableRememberMeValues[0].(string)
  if strconv.FormatBool(securitySettings.DisableSignup) != disableRememberMe {
    diff += greenString("+ disable_remember_me " + strconv.FormatBool(securitySettings.DisableSignup)) + redString("\n- disable_remember_me " + disableRememberMe) + "\n"
  }


  if diff != "" {
    diff = "security\n" + diff
  }

  return diff, nil

}

func handleGeneralSettingsDryRun(client *ssh.Client, generalSettings *General) (string, error) {

	currentConfigStr := readFile("/var/lib/jenkins/config.xml", client)
	mv, _ := mxj.NewMapXml([]byte(currentConfigStr))
  values, err := mv.ValuesForKey("numExecutors")
  if err != nil {
    fmt.Println("Couldn't get NumExecutors")
    return "", err
  }

  diff := ""

  numExecutors := values[0].(string)
  if strconv.Itoa(generalSettings.NumExecutors) != numExecutors {
    diff += greenString("+ num_executors " + strconv.Itoa(generalSettings.NumExecutors)) + redString("\n- num_executors " + numExecutors) + "\n"
  }


  workspaceDirValues, err := mv.ValuesForKey("workspaceDir")
  if err != nil {
    fmt.Println("Couldn't get workspaceDir")
    return "", err
  }

  workspaceDir := workspaceDirValues[0].(string)
  if generalSettings.WorkspaceDir != workspaceDir {
    diff += greenString("+ workspace_dir " + generalSettings.WorkspaceDir) + redString("\n- workspace_dir " + workspaceDir) + "\n"
  }

  if diff != "" {
    diff = "general\n" + diff
  }

  return diff, nil
	//  mv.UpdateValuesForPath("numExecutors:4", "hudson")
  //	xmlValue, _ := mv.Xml()
  //	fmt.Println(string(xmlValue))
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
  }
/*
	session, err := client.NewSession()

	if err != nil {
		log.Fatalln("Failed to create session: " + err.Error())
	}

	dest := "/home/sidshanker/blah/blablah/htxt.txt"

	directoryName := filepath.Dir(dest)

	mkdirCommand := shellquote.Join("sudo", "mkdir", "-p", directoryName)

  fmt.Println("Running mkdir")
	runShellCommand(mkdirCommand, client)
  fmt.Println("About to scp")
	err = copyPath(f.Name(), dest, session)
	if err != nil {
		log.Fatalln("Failed to scp: " + err.Error())
	}

	if _, err := os.Stat(f.Name()); os.IsNotExist(err) {
		fmt.Printf("no such file or directory: %s", dest)
	}
  */
}

func init() {
	// We pass the user variable we declared at the package level (above).
	// The "&" character means we are passing the variable "by reference" (as opposed to "by value"),
	// meaning: we don't want to pass a copy of the user variable. We want to pass the original variable.
	flag.BoolVarP(&dryRun, "dry-run", "d", false, "Run tuxedo in dry-run mode")
}
