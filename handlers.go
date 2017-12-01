package main

import (
  "fmt"
  "strconv"
	"golang.org/x/crypto/ssh"
	"github.com/clbanning/mxj"
)

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
  if strconv.FormatBool(securitySettings.DisableRememberMe) != disableRememberMe {
    diff += greenString("+ disable_remember_me " + strconv.FormatBool(securitySettings.DisableSignup)) + redString("\n- disable_remember_me " + disableRememberMe) + "\n"
  }


  if diff != "" {
    diff = "security\n" + diff
  }

  return diff, nil

}

func handleSecuritySettings(client *ssh.Client, securitySettings *Security)  error {
	currentConfigStr := readFile("/var/lib/jenkins/config.xml", client)
	mv, _ := mxj.NewMapXml([]byte(currentConfigStr))

  disableSignupValues, err := mv.ValuesForKey("disableSignup")
  if err != nil {
    fmt.Println("Couldn't get disableSignup")
    return err
  }

  disableSignup := disableSignupValues[0].(string)
  if strconv.FormatBool(securitySettings.DisableSignup) != disableSignup {
    mv.UpdateValuesForPath("disableSignup" + ":" + strconv.FormatBool(securitySettings.DisableSignup), "hudson.securityRealm")
  }

  disableRememberMeValues, err := mv.ValuesForKey("disableRememberMe")
  if err != nil {
    fmt.Println("Couldn't get disableRememberMe")
    return err
  }

  disableRememberMe := disableRememberMeValues[0].(string)
  if strconv.FormatBool(securitySettings.DisableRememberMe) != disableRememberMe {
    mv.UpdateValuesForPath("disableRememberMe" + ":" + strconv.FormatBool(securitySettings.DisableRememberMe), "hudson")
  }

  xmlValue, _ := mv.Xml()

	dest := "/var/lib/jenkins/config.xml"
  return scpFileToServer(client, string(xmlValue), dest)
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
}

func handleGeneralSettings(client *ssh.Client, generalSettings *General)  error {
	currentConfigStr := readFile("/var/lib/jenkins/config.xml", client)
	mv, _ := mxj.NewMapXml([]byte(currentConfigStr))
  values, err := mv.ValuesForKey("numExecutors")
  if err != nil {
    fmt.Println("Couldn't get NumExecutors")
    return err
  }

  numExecutors := values[0].(string)
  if strconv.Itoa(generalSettings.NumExecutors) != numExecutors {
    mv.UpdateValuesForPath("numExecutors" + ":" + strconv.Itoa(generalSettings.NumExecutors), "hudson")
  }

  workspaceDirValues, err := mv.ValuesForKey("workspaceDir")
  if err != nil {
    fmt.Println("Couldn't get workspaceDir")
    return err
  }

  workspaceDir := workspaceDirValues[0].(string)
  if generalSettings.WorkspaceDir != workspaceDir {
    mv.UpdateValuesForPath("workspaceDir" + ":" + generalSettings.WorkspaceDir, "hudson")
  }

  xmlValue, _ := mv.Xml()

	dest := "/var/lib/jenkins/config.xml"
  return scpFileToServer(client, string(xmlValue), dest)
}
