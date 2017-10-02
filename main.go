/*
Package healthcheck implements a web server that returns "healthy:true" or "healthy:false"

The value of 'healthy' is determined by how the service is started.

The syntax to start the service is:

    ./healthcheck (1 or more commands to execute- separated by spaces)*
      * commands with multiple args need to be quoted

If all passed bash commands execute with exit code 0- the health check returns "true"
If any of the bash commands exit with a non-zero exit code, the health check returns "false"

   IE-  ./healthcheck "docker ps"  would bind the value of "healthy:<bool>" to the result of running "docker ps"

        http://hostname:8080/health

        returns:     healthy:true
                     version:17.06.0-ce-rc4
*/
package main

import (
	"net/http"
	"os"
	"os/exec"
	"fmt"
	"errors"
	"time"
)

// Define Timeout Variables
const (
	timeoutDur = 5 * time.Second
	timeoutMsg = "Request timed out."
)

// Create http handler
type Handler struct{}

// http handler method (ServeHttp)
func ( *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var version, health string
	var err, err2 error

	// See if service is healthy based on bash commands executed
	health, err = isHealthy()
	// Handle a non zero exit code
	if err != nil {
		err = errors.New(health)
		health = "false"
	}
	// Execute bash command to get desired "version" output- docker version for us here
	version, err2 = runBash("docker version --format '{{.Server.Version}}'")
	// Handle errors with obtaining version (non zero exit code)
	if err2 != nil {
		version = "error:" + version
	}
	// Format and return response based on health and version outputs
	response := httpResponse(health, version , err)
	fmt.Fprintf(w, response)
}

/* Returns "true" if all bash commands given as args exit with 0
   Returns "false" if any bash commands given as args exit with non-zero */
func isHealthy() (result string, err error) {

	args := os.Args[1:]

	for count := 0; count < len(args); count++ {
		out, err := runBash(args[count])
		if err != nil {
			return out, err
		}
	}
	return "true", nil

}
// Formats the http response
func httpResponse (output string, version string, e error) (string){

	var webout string

	webout =
		"healthy:" + output + "\n" +
			"version:" + version + "\n"

	if e != nil {
		webout = webout + "\nError: " + e.Error()
	}
	return webout
}

// Takes a single bash command, runs it, and returns stdout or stderr along with any error.
func runBash(command string) (stdout string, e error) {

	out, err := exec.Command("bash", "-c", command).CombinedOutput()

	if err != nil {
		return string(out[5:]), errors.New(string(out[:]))
	}
	return string(out), err
}

// Main
func main() {

	indexHandler := &Handler{}

	http.Handle("/", http.TimeoutHandler(indexHandler, timeoutDur, timeoutMsg))
	http.ListenAndServe(":8080", nil)
}