/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package docker

import docker "github.com/fsouza/go-dockerclient"

type Client interface {
	CreateContainer(opts docker.CreateContainerOptions) (*docker.Container, error)
	StartContainer(id string, hostConfig *docker.HostConfig) error
	Logs(opts docker.LogsOptions) error
	ListImages(opts docker.ListImagesOptions) ([]docker.APIImages, error)
	ListContainers(opts docker.ListContainersOptions) ([]docker.APIContainers, error)
	CreateExec(opts docker.CreateExecOptions) (*docker.Exec, error)
	StartExecNonBlocking(id string, opts docker.StartExecOptions) (docker.CloseWaiter, error)
	StartExec(id string, opts docker.StartExecOptions) error
	InspectContainer(id string) (*docker.Container, error)
	AttachToContainerNonBlocking(opts docker.AttachToContainerOptions) (docker.CloseWaiter, error)
	AttachToContainer(opts docker.AttachToContainerOptions) error
	RemoveContainer(opts docker.RemoveContainerOptions) error
	PullImage(opts docker.PullImageOptions, auth docker.AuthConfiguration) error
	InspectExec(id string) (*docker.ExecInspect, error)

	Version() (*docker.Env, error)
	Info() (*docker.DockerInfo, error)
}
