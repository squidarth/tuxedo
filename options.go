package main

type Config struct {
  SSHSettingsConfig  SSHSettings `hcl:"ssh_settings"`
  SecurityConfig  Security `hcl:"security"`
  GeneralConfig  General `hcl:"general"`
}

type SSHSettings struct {
  HostIp      string `hcl:"host_ip"`
  SSHUsername string  `hcl:"ssh_username"`
  PathToKey   string  `hcl:"path_to_key"`
}

type Security struct {
  DisableSignup     bool `hcl:"disable_signup"`
  DisableRememberMe bool `hcl:"disable_remember_me"`
}

type General struct {
  JenkinsDir   string `hcl:"jenkins_dir"`
  NumExecutors int    `hcl:"num_executors"`
  WorkspaceDir string `hcl:"workspace_dir"`
}

//type Plugin struct {
//  name string
//}
