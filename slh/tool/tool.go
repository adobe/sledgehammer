/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package tool

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/adobe/sledgehammer/slh/config"
	"github.com/adobe/sledgehammer/utils"
	secrets "github.com/adobe/sledgehammer/utils/docker"
	bolt "github.com/coreos/bbolt"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/sirupsen/logrus"
)

var (
	// ErrorToolNotFound will be thrown when the tool was not found
	ErrorToolNotFound = errors.New("Tool not found")
	// ErrorRegistryEmpty will be thrown when the registry is empty
	ErrorRegistryEmpty = errors.New("Registry cannot be empty")
	// ErrorNameEmpty will be thrown if the name of the tool is empty
	ErrorNameEmpty = errors.New("Name cannot be empty")
	// ErrorNameInvalid will be thrown if the name of the tool is inavlid
	ErrorNameInvalid = errors.New("Name is not valid")
	// ErrorTypeInvalid will be thrown if the Type of the tool is inavlid
	ErrorTypeInvalid = errors.New("Type is not valid")
	// NameRegex is the pattern that all tool names must match
	NameRegex   = "^[a-zA-Z0-9-_]+$"
	restoreFunc = func(state *terminal.State, buffer *bytes.Buffer, out io.Writer) error {
		if state != nil {
			logrus.Infoln("Restoring terminal to original state")
			err := terminal.Restore(int(os.Stdin.Fd()), state)
			logrus.SetOutput(out)
			logrus.Infoln("Sending buffered logs")
			scanner := bufio.NewScanner(buffer)
			for scanner.Scan() {
				text := string(scanner.Bytes())
				if len(text) > 0 {
					fmt.Fprintln(out, text)
				}
			}
			if err != nil {
				return err
			}
		}
		return nil
	}
)

// JSON is the structure that will be stored in the database
type JSON struct {
	Type string          `json:"type"`
	Tool json.RawMessage `json:"tool"`
}

// Factory is a factory to create tools from the db and from user input
type Factory interface {
	Raw() Tool
	Create(Data) Tool
}

// Tool is the base interface for a tool and can be used to support multiple tools
type Tool interface {
	// Versions will return all version that are currently available for the given tool
	Versions() ([]string, error)
	Data() *Data
}

// Tools is the struct that has access to all tools in the database
type Tools struct {
	config.Database
}

// ExecutionOptions are used to execute commands on a container
type ExecutionOptions struct {
	Tool      Tool
	Version   string
	IO        *config.IO
	Docker    *config.Docker
	Arguments []string
	Mounts    []string
}

var (
	// BucketKey is the key under which all tools are stored in the database
	BucketKey = "tools"
	// Types define the possible types of a tool
	Types = map[string]Factory{
		"local": &LocalFactory{},
		"hub":   &HubFactory{},
		"jfrog": &JFrogFactory{},
		"":      &HubFactory{},
	}
)

// New will instantiate a new Tools struct that can be used to access all tools registered with Sledgehammer7
// The structure of the tool will look like this:
// tools
// |- foo
// |  |- default (foo tool in the default registry)
// |  |- other (foo tool in the other registry)
// |- bar
//    |- default/bar (bar tool in the default registry)
// For a registry only unique tools are allowed.
// In the case that there are multiple entries with the same name in the same registry
// it is up to the registry to select the one with the highest priority.
func New(db config.Database) *Tools {
	return &Tools{
		Database: db,
	}
}

// From will return all tools from a given registry name
func (t *Tools) From(registry string) ([]Tool, error) {
	tools := []Tool{}
	logrus.WithField("registry", registry).Debug("Fetching all tools from the registry")
	_, toolsMap, err := t.List()
	if err != nil {
		return tools, err
	}
	for _, v := range toolsMap {
		for _, tool := range v {
			if tool.Data().Registry == registry {
				tools = append(tools, tool)
			}
		}
	}
	return tools, nil
}

// Remove will remove the given tool from the list of tools registered with Sledgehammer
func (t *Tools) Remove(registry string, tool string) error {
	if len(registry) == 0 {
		return ErrorRegistryEmpty
	}
	if len(tool) == 0 {
		return ErrorNameEmpty
	}
	logrus.WithField("registry", registry).WithField("tool", tool).Debug("Removing tool from registry")
	err := t.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BucketKey))
		if err != nil {
			return err
		}
		// get tool bucket
		if bucket != nil {
			// get bucket of the given tool
			toolBucket := bucket.Bucket([]byte(tool))
			if toolBucket != nil {
				// there is at least one tool available that we can remove
				size := toolBucket.Stats().KeyN
				//  there is more than one tool...
				if size > 1 {
					// get the tool to remove
					logrus.Debug("Found more than one tool")
					toolJSON := toolBucket.Get([]byte(registry))
					if toolJSON != nil {
						var m JSON
						err := json.Unmarshal(toolJSON, &m)
						if err != nil {
							return err
						}
						fun, found := Types[m.Type]
						if found {
							dbTool := fun.Raw()
							err := json.Unmarshal(m.Tool, dbTool)
							if err != nil {
								return err
							}
							// delete the tool from the bucket
							err = toolBucket.Delete([]byte(registry))
							if err != nil {
								return err
							}
							// If it was the default tool
							if dbTool.Data().Default {
								logrus.WithField("tool", tool).Debug("Default tool deleted, need to select a new one")
								// select a new default
								// get oldest tool
								timeToBeat := time.Now()
								var oldestTool Tool

								toolBucket.ForEach(func(_ []byte, value []byte) error {
									var m JSON
									err := json.Unmarshal(value, &m)
									if err != nil {
										return err
									}
									// Depending on the type, we can run json.Unmarshal again on the same byte slice
									// But this time, we'll pass in the appropriate struct instead of a map
									fun, found := Types[m.Type]
									if found {
										dbTool = fun.Raw()
										err := json.Unmarshal(m.Tool, dbTool)
										if err != nil {
											return err
										}
										if dbTool.Data().Added.Before(timeToBeat) {
											timeToBeat = dbTool.Data().Added
											oldestTool = dbTool
										}
									}
									return nil
								})

								oldestTool.Data().Default = true
								logrus.WithField("tool", oldestTool.Data().Registry+"/"+oldestTool.Data().Name).Debug("Selected a new default tool")
								// write again
								jsonTool, err := json.Marshal(oldestTool)
								if err != nil {
									return err
								}
								bb, err := json.Marshal(JSON{Tool: jsonTool, Type: oldestTool.Data().Type})
								if err != nil {
									return err
								}
								return toolBucket.Put([]byte(oldestTool.Data().Registry), bb)
							}
						}
					}
				} else {
					// we only have one tool
					err := bucket.DeleteBucket([]byte(tool))
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	return err
}

// Add will add the given tool to Sledgehammer
func (t *Tools) Add(toolsToAdd ...Tool) error {
	for _, tool := range toolsToAdd {
		if len(tool.Data().Name) == 0 {
			return ErrorNameEmpty
		}
		if len(tool.Data().Registry) == 0 {
			return ErrorRegistryEmpty
		}
		if _, found := Types[tool.Data().Type]; !found {
			return ErrorTypeInvalid
		}
	}

	err := t.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BucketKey))
		if err != nil {
			return err
		}
		if bucket != nil {
			for _, tool := range toolsToAdd {
				logrus.WithField("tool", tool.Data().Name).Debug("Adding new tool")
				toolBucket, err := bucket.CreateBucketIfNotExists([]byte(tool.Data().Name))
				if err != nil {
					return err
				}
				if toolBucket != nil {
					// set default if there is no tool yet
					isDefault := toolBucket.Stats().KeyN == 0
					update := toolBucket.Get([]byte(tool.Data().Registry))
					// tool already exists
					if update != nil {
						// exist, update
						var m JSON
						err := json.Unmarshal(update, &m)
						if err != nil {
							return err
						}
						// Depending on the type, we can run json.Unmarshal again on the same byte slice
						// But this time, we'll pass in the appropriate struct instead of a map
						fun, found := Types[m.Type]
						if found {
							logrus.WithField("tool", tool.Data().Name).Debug("Tool found, updating")
							dbTool := fun.Raw()
							err := json.Unmarshal(m.Tool, dbTool)
							if err != nil {
								return err
							}
							// set options and insert again
							tool.Data().Default = dbTool.Data().Default
							tool.Data().Added = dbTool.Data().Added

							jsonTool, err := json.Marshal(tool)
							if err != nil {
								return err
							}
							bb, err := json.Marshal(JSON{Tool: jsonTool, Type: tool.Data().Type})
							if err != nil {
								return err
							}
							err = toolBucket.Put([]byte(tool.Data().Registry), bb)
							if err != nil {
								return err
							}
						}
					} else {
						// new, add
						logrus.WithField("tool", tool.Data().Name).Debug("Tool is new, adding")
						tool.Data().Default = isDefault
						tool.Data().Added = time.Now()
						jsonTool, err := json.Marshal(tool)
						if err != nil {
							return err
						}
						bb, err := json.Marshal(JSON{Tool: jsonTool, Type: tool.Data().Type})
						if err != nil {
							return err
						}
						err = toolBucket.Put([]byte(tool.Data().Registry), bb)
						if err != nil {
							return err
						}
					}
				}
			}
		}
		return nil
	})
	return err
}

// Search will return a list filtered by the search term. A simple contains is supported, nothing else.
func (t *Tools) Search(search string) ([]string, map[string][]Tool, error) {
	newSorted := []string{}
	newMap := map[string][]Tool{}
	sorted, ts, err := t.List()
	if err != nil {
		return newSorted, newMap, err
	}
	for _, v := range sorted {
		if strings.Contains(v, search) {
			newSorted = append(newSorted, v)
			newMap[v] = ts[v]
		}
	}
	return newSorted, newMap, nil
}

// List will list all tools currently stored in the database and returns a sorted list of toolnames and the tools
func (t *Tools) List() ([]string, map[string][]Tool, error) {
	logrus.Debug("Listing all tools")
	tools := map[string][]Tool{}
	sortedTools := []string{}

	err := t.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketKey))
		if bucket != nil {
			// get all tool buckets
			bucket.ForEach(func(key, _ []byte) error {
				// this is the tools bucket, get all tools
				toolBucket := bucket.Bucket(key)
				if toolBucket != nil {
					sortedTools = append(sortedTools, string(key))
					// loop over all tools
					tools[string(key)] = []Tool{}
					toolBucket.ForEach(func(registry, toolJSON []byte) error {
						var m JSON
						err := json.Unmarshal(toolJSON, &m)
						if err != nil {
							return err
						}
						fun, found := Types[m.Type]
						if found {
							tool := fun.Raw()
							err := json.Unmarshal(m.Tool, tool)
							if err != nil {
								return err
							}
							logrus.WithField("tool", tool.Data().Name).Debug("Found tool")
							tools[string(key)] = append(tools[string(key)], tool)
						}
						return nil
					})
				}
				return nil
			})
		}
		return nil
	})
	// sort the tools alphabetically
	sort.Strings(sortedTools)
	return sortedTools, tools, err
}

// Versions will get all local version that are available for the given tool
func Versions(client config.Docker, to Tool) ([]string, error) {
	versions := []string{}
	logrus.WithField("filter", to.Data().Image).Info("Checking local images")
	images, err := client.Docker.ListImages(docker.ListImagesOptions{
		Filter: fullImageName(to),
	})
	if err != nil {
		return versions, err
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			versions = append(versions, strings.Replace(tag, FullImage(to, "")+":", "", 1))
		}
	}
	logrus.WithField("version", versions).Debug("Found local images with tags")
	return versions, err
}

// Get will return a single tool from a registry
func (t *Tools) Get(registry string, name string) (Tool, error) {
	logrus.WithField("registry", registry).WithField("tool", name).Debug("Getting tool from registry")
	var selectedTool Tool
	err := t.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BucketKey))
		if err != nil {
			return err
		}
		if bucket != nil {
			toolBucket := bucket.Bucket([]byte(name))
			if toolBucket != nil {
				value := toolBucket.Get([]byte(registry))
				if value != nil {
					var m JSON
					err := json.Unmarshal(value, &m)
					if err != nil {
						return err
					}
					fun, found := Types[m.Type]
					if found {
						tool := fun.Raw()
						err := json.Unmarshal(m.Tool, tool)
						if err != nil {
							return err
						}
						selectedTool = tool
					}
				}
				if registry == "" {
					// get default tool
					toolBucket.ForEach(func(_, toolJSON []byte) error {
						var m JSON
						err := json.Unmarshal(toolJSON, &m)
						if err != nil {
							return err
						}
						fun, found := Types[m.Type]
						if found {
							to := fun.Raw()
							err := json.Unmarshal(m.Tool, to)
							if err != nil {
								return err
							}
							if to.Data().Default {
								logrus.WithField("tool", to.Data().Name).WithField("registry", to.Data().Registry).WithField("type", to.Data().Type).Debug("Found tool")
								selectedTool = to
								return nil
							}
						}
						return nil
					})
				}
			}
		}
		return nil
	})
	if err != nil {
		return selectedTool, err
	}
	if selectedTool != nil {
		return selectedTool, nil
	}
	return selectedTool, ErrorToolNotFound
}

// Pull will try to pull the given tool with the given version from the remote repository
func Pull(client config.Docker, to Tool, tag string, timeout time.Duration) error {

	doneChan := make(chan error, 1)

	dockerCreds := docker.AuthConfiguration{}

	creds, _ := secrets.GetCredentials(to.Data().ImageRegistry)
	if creds != nil {
		dockerCreds.ServerAddress = creds.ServerURL
		dockerCreds.Username = creds.Username
		dockerCreds.Password = creds.Secret
	}

	ctx, can := context.WithTimeout(context.Background(), timeout)
	defer can()

	pr, pw := io.Pipe()
	defer pw.Close()
	defer pr.Close()

	scanner := bufio.NewScanner(pr)
	go func() {
		for scanner.Scan() {
			logrus.Info(scanner.Text())
		}
	}()

	logrus.WithFields(logrus.Fields{
		"registry":   to.Data().ImageRegistry,
		"repository": to.Data().Image,
		"tag":        tag,
	}).Infoln("Pulling image")

	pollFunc := func(doneChan chan error) {
		err := client.Docker.PullImage(docker.PullImageOptions{
			// legacy, needed for old clients
			Registry:     to.Data().ImageRegistry,
			Repository:   FullImage(to, ""),
			Tag:          tag,
			OutputStream: pw,
			Context:      ctx,
		}, dockerCreds)
		doneChan <- err
	}
	go pollFunc(doneChan)
	select {
	case err := <-doneChan:
		return err
	case <-ctx.Done():
		return errors.New("Timeout reached during image polling")
	}
}

// StartIfDaemon will start the given tool if it is a daemon and will return the id of the prepared container.
// If no id is returned, then the tool is no daemon
func StartIfDaemon(opt *ExecutionOptions) (string, error) {
	if opt.Tool.Data().Daemon == nil {
		logrus.Info("No daemon tool detected")
		return "", nil
	}
	logrus.WithFields(logrus.Fields{
		"image":     opt.Tool.Data().Image,
		"tool":      opt.Tool.Data().Name,
		"arguments": opt.Tool.Data().Daemon.Entry,
	}).Info("Detected daemon tool")

	conf := &docker.Config{
		Image:        FullImage(opt.Tool, opt.Version),
		Entrypoint:   opt.Tool.Data().Daemon.Entry,
		Env:          utils.PrepareEnvironment(os.Environ()),
		AttachStderr: false,
		AttachStdout: false,
		AttachStdin:  false,
		Tty:          true,
		OpenStdin:    false,
	}

	resp, err := opt.Docker.Docker.CreateContainer(docker.CreateContainerOptions{
		Config: conf,
		HostConfig: &docker.HostConfig{
			GroupAdd:    []string{"0"},
			NetworkMode: "host",
			AutoRemove:  true,
			Mounts:      utils.PrepareMounts(opt.Mounts),
		},
	})

	if err != nil {
		return "", err
	}
	if err := opt.Docker.Docker.StartContainer(resp.ID, nil); err != nil {
		return "", err
	}

	return resp.ID, nil
}

// Execute will execute the given command in an already running container.
func Execute(containerID string, opt *ExecutionOptions) (int, error) {
	isPipe := utils.IsPipe()
	stdOut := &bytes.Buffer{}

	var state *terminal.State
	var err error

	workspace, err := utils.WorkingDirectory(opt.Mounts)
	if err != nil {
		return 1, err
	}

	arguments := append(opt.Tool.Data().Entry, opt.Arguments...)
	logrus.WithField("arguments", strings.Join(arguments, ",")).Info("Prepared arguments for exec")

	createExecConfig := docker.CreateExecOptions{
		Container:    containerID,
		AttachStderr: !isPipe,
		AttachStdout: !isPipe,
		AttachStdin:  !isPipe,
		Cmd:          arguments,
		Tty:          !isPipe,
		Env:          utils.PrepareEnvironment(os.Environ()),
		WorkingDir:   workspace,
	}
	if os.Getuid() >= 0 && os.Getgid() >= 0 {
		createExecConfig.User = strconv.Itoa(os.Getuid()) + ":" + strconv.Itoa(os.Getgid())
	}

	inP, outP, errP := preparePipes(opt.IO)
	execConfig := docker.StartExecOptions{
		ErrorStream:  errP,
		InputStream:  inP,
		OutputStream: outP,
		Tty:          !isPipe,
	}
	if !isPipe {
		execConfig.RawTerminal = true
	}

	logrus.Debugln("Creating Exec")
	exec, err := opt.Docker.Docker.CreateExec(createExecConfig)
	if err != nil {
		logrus.WithField("tool", opt.Tool.Data().Name).Errorln("Could not get logs from tool container: ", err.Error())
		return 1, err
	}

	tty := terminal.IsTerminal(int(os.Stdin.Fd()))
	if tty && os.Stdin == opt.IO.In {
		logrus.Info("Detected tty terminal, making it raw")
		// buffer concurrent output, it does not work well with a raw terminal...
		logrus.SetOutput(stdOut)
		state, err = terminal.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return 1, err
		}
	}

	logrus.Debugln("Starting Exec")
	err = opt.Docker.Docker.StartExec(exec.ID, execConfig)

	if err != nil {
		restoreFunc(state, stdOut, opt.IO.Out)
		logrus.WithField("tool", opt.Tool.Data().Name).Errorln("Could not get logs from tool container: ", err.Error())
		return 1, err
	}
	err = restoreFunc(state, stdOut, opt.IO.Out)
	if err != nil {
		return 1, err
	}

	// get execute information for the exit code
	execCon, err := opt.Docker.Docker.InspectExec(exec.ID)
	if err != nil {
		return 1, err
	}
	return execCon.ExitCode, nil
}

// StartAndExecute will start a container and execute the given command.
// When this method is called, the tool is not a daemon tool
func StartAndExecute(opt *ExecutionOptions) (int, error) {
	var state *terminal.State
	isPipe := utils.IsPipe()

	stdOut := &bytes.Buffer{}

	workspace, err := utils.WorkingDirectory(opt.Mounts)
	if err != nil {
		return 1, err
	}

	conf := &docker.Config{
		Image:        FullImage(opt.Tool, opt.Version),
		Cmd:          opt.Arguments,
		Env:          utils.PrepareEnvironment(os.Environ()),
		WorkingDir:   workspace,
		AttachStderr: !isPipe,
		AttachStdout: !isPipe,
		AttachStdin:  !isPipe,
		Tty:          !isPipe,
		OpenStdin:    true,
		StdinOnce:    isPipe,
	}
	if os.Getuid() >= 0 && os.Getgid() >= 0 {
		conf.User = strconv.Itoa(os.Getuid()) + ":" + strconv.Itoa(os.Getgid())
	}
	if len(opt.Tool.Data().Entry) > 0 {
		conf.Entrypoint = opt.Tool.Data().Entry
	}
	logrus.Info("Creating container")
	resp, err := opt.Docker.Docker.CreateContainer(docker.CreateContainerOptions{
		Config: conf,
		HostConfig: &docker.HostConfig{
			GroupAdd:    []string{"0"},
			NetworkMode: "host",
			AutoRemove:  true,
			Mounts:      utils.PrepareMounts(opt.Mounts),
		},
	})

	if err != nil {
		return 1, err
	}
	inP, outP, errP := preparePipes(opt.IO)
	attachConfig := docker.AttachToContainerOptions{
		Container:    resp.ID,
		Logs:         true,
		Stderr:       true,
		Stdout:       true,
		Stdin:        true,
		Stream:       true,
		ErrorStream:  errP,
		InputStream:  inP,
		OutputStream: outP,
	}
	if !isPipe {
		attachConfig.RawTerminal = true
	}

	if err := opt.Docker.Docker.StartContainer(resp.ID, nil); err != nil {
		logrus.WithField("tool", opt.Tool.Data().Name).Errorln("Could not start tool container")
		return 1, err
	}

	tty := terminal.IsTerminal(int(os.Stdin.Fd()))
	if tty && os.Stdin == opt.IO.In {
		logrus.Info("Detected tty terminal, making it raw")
		// buffer concurrent output, it does not work well with a raw terminal...
		logrus.SetOutput(stdOut)
		state, err = terminal.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return 1, err
		}
	}

	logrus.Debugln("Attaching to container")
	err = opt.Docker.Docker.AttachToContainer(attachConfig)
	logrus.Debugln("Done attaching to container")

	if err != nil {
		restoreFunc(state, stdOut, opt.IO.Out)
		logrus.WithField("tool", opt.Tool.Data().Name).Errorln("Could not get logs from tool container")
		return 1, err
	}

	logrus.Debugln("Done...")
	err = restoreFunc(state, stdOut, opt.IO.Out)
	if err != nil {
		return 1, err
	}

	// get container information for the exit code
	cont, err := opt.Docker.Docker.InspectContainer(resp.ID)
	if err != nil {
		return 1, err
	}

	return cont.State.ExitCode, nil
}

// StopDaemons will stop all tool daeomons if possible
func StopDaemons() {}

func fullImageName(to Tool) string {
	if len(to.Data().ImageRegistry) > 0 {
		return to.Data().ImageRegistry + "/" + to.Data().Image
	}
	return to.Data().Image
}

// HasValidName will check if the name of the tool is valid according to the regex
func HasValidName(name string) bool {
	valid, _ := regexp.MatchString(NameRegex, name)
	return valid
}

// Exists will check if a given tool exists in the collection of tools
func Exists(tools []Tool, tool Tool) bool {
	for _, to := range tools {
		if to.Data().Name == tool.Data().Name &&
			to.Data().Registry == tool.Data().Registry {
			return true
		}
	}
	return false
}

func preparePipes(cfg *config.IO) (io.Reader, io.Writer, io.Writer) {
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	errR, errW := io.Pipe()

	go func() {
		defer inW.Close()
		io.Copy(inW, cfg.In)
	}()
	go func() {
		defer outW.Close()
		io.Copy(cfg.Out, outR)
	}()
	go func() {
		defer errW.Close()
		io.Copy(cfg.Err, errR)
	}()

	return inR, outW, errW
}
