package main

import (
	"bytes"
	"fmt"
	"github.com/clbanning/mxj"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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

	cmd := shellquote.Join("scp", "-t", destination)
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

func main() {
	flag.Parse()
	if flag.NArg() > 0 {
		flag.Usage()
		os.Exit(1)
	}
	f, _ := ioutil.TempFile("", "")
	fmt.Fprintln(f, "hello world")
	f.Close()
	//	defer os.Remove(f.Name())
	//	defer os.Remove(f.Name() + "-copy")

	//	agent, err := getAgent()
	//	if err != nil {
	//		log.Fatalln("Failed to connect to SSH_AUTH_SOCK:", err)
	//	}

	privateKey, err := ioutil.ReadFile("/Users/sidharthshanker/.ssh/id_rsa")
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		fmt.Println("Failed to parse the private key file")
		fmt.Println(err.Error())
	}

	client, err := ssh.Dial("tcp", "35.193.209.57:22", &ssh.ClientConfig{
		User: os.Getenv("USER"),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // FIXME: please be more secure in checking host keys
	})
	if err != nil {
		log.Fatalln("Failed to dial:", err)
	}

	currentConfigStr := readFile("/var/lib/jenkins/config.xml", client)

	mv, _ := mxj.NewMapXml([]byte(currentConfigStr))

	mv.UpdateValuesForPath("numExecutors:4", "hudson")
	//	xmlValue, _ := mv.Xml()
	//	fmt.Println(string(xmlValue))
	session, err := client.NewSession()

	if err != nil {
		log.Fatalln("Failed to create session: " + err.Error())
	}

	dest := "/home/sidharthshanker/blah/blablah/htxt.txt"

	directoryName := filepath.Dir(dest)

	mkdirCommand := shellquote.Join("mkdir", "-p", directoryName)
	runShellCommand(mkdirCommand, client)
	err = copyPath(f.Name(), dest, session)
	if err != nil {
		log.Fatalln("Failed to scp: " + err.Error())
	}

	if _, err := os.Stat(f.Name()); os.IsNotExist(err) {
		fmt.Printf("no such file or directory: %s", dest)
	}
	fmt.Printf("dryRun value: %t\n", dryRun)
}

func init() {
	// We pass the user variable we declared at the package level (above).
	// The "&" character means we are passing the variable "by reference" (as opposed to "by value"),
	// meaning: we don't want to pass a copy of the user variable. We want to pass the original variable.
	flag.BoolVarP(&dryRun, "dry-run", "d", false, "Run tuxedo in dry-run mode")
}
