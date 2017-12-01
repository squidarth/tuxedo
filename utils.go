package main

import (
  "bytes"
  "fmt"
	"os"
	"golang.org/x/crypto/ssh"
	"path"
	"io"
  "log"
	"io/ioutil"
	shellquote "github.com/kballard/go-shellquote"
)

func scpFileToServer(client *ssh.Client, content string, dest string) error {
  f, _ := ioutil.TempFile("", "")
  fmt.Fprintln(f, content)
  f.Close()

  defer os.Remove(f.Name())
  defer os.Remove(f.Name() + "-copy")
	session, err := client.NewSession()

	if err != nil {
		log.Fatalln("Failed to create session: " + err.Error())
	}


	err = copyPath(f.Name(), dest, session)
	if err != nil {
		log.Fatalln("Failed to scp: " + err.Error())
	}

  return nil
}


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

	errors := make(chan error) //  Creates channel over which the goroutine can communicate

	go func() {
		errors <- session.Wait() // write to errors
	}()

	fmt.Fprintf(w, "C%#o %d %s\n", mode, size, fileName)
	_, copyErr := io.Copy(w, contents)
	if copyErr != nil {
		fmt.Println(copyErr.Error())
	}
	fmt.Fprint(w, "\x00")
	w.Close()

	return <-errors // receive from errors
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
