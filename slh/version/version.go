/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package version

import (
	"sort"

	"github.com/Masterminds/semver"
	"github.com/sirupsen/logrus"
)

var (
	// Version represents the semantic version of Sledgehammer
	Version string = "0.0.1"
	// BuildDate represents the date of the build
	BuildDate string
	// GitCommit represent the git sha of the current build
	GitCommit string
	// DefaultConstraint is the constraint used for version selection when none is given
	DefaultConstraint = "*"
)

// Merge will return a merged sorted list of the two lists a and b
func Merge(a []string, b []string) []string {
	set := make(map[string]struct{})
	for _, val := range a {
		set[val] = struct{}{}
	}
	for _, val := range b {
		set[val] = struct{}{}
	}
	versions := []string{}
	for key := range set {
		versions = append(versions, key)
	}
	sort.Strings(versions)
	return versions
}

// Has will check if a list of versions contain the given version
func Has(versions []string, version string) bool {
	for _, val := range versions {
		if val == version {
			return true
		}
	}
	return false
}

// Select will select one of the given versions based on the given constraint. If no version can be found, it will return an empty string
func Select(versions []string, constraint string) string {
	if constraint == "" {
		constraint = DefaultConstraint
	}
	parsedConstraint, err := semver.NewConstraint(constraint)
	if err != nil {
		if Has(versions, constraint) {
			return constraint
		}
		logrus.WithFields(logrus.Fields{
			"versions":   versions,
			"constraint": constraint}).Infoln("Could not find a version")
		return ""
	}
	// if we are here then all versions are sem ver compatible
	sort.Strings(versions)
	parsedVersions := []*semver.Version{}
	for i := len(versions) - 1; i >= 0; i-- {
		ver := versions[i]
		parsedVersion, err := semver.NewVersion(ver)
		if err != nil {
			continue
		}
		parsedVersions = append(parsedVersions, parsedVersion)
	}
	sort.Sort(semver.Collection(parsedVersions))
	for i := len(parsedVersions) - 1; i >= 0; i-- {
		ver := parsedVersions[i]
		matches := parsedConstraint.Check(ver)
		if matches {
			return ver.Original()
		}
	}
	if Has(versions, "latest") {
		logrus.WithFields(logrus.Fields{
			"versions":   versions,
			"constraint": constraint,
			"selected":   "latest"}).Infoln("Selected version")
		return "latest"
	}
	return ""
}

// ShouldPull will determine if the remote version should be pulled.
// It will be pulled if the remote version is newer than the local one.
func ShouldPull(local string, remote string) bool {
	localSem, err1 := semver.NewVersion(local)
	remoteSem, err2 := semver.NewVersion(remote)

	logrus.WithField("remote", remote).WithField("local", local).Debug("Checking version for pulling")

	if err1 != nil || err2 != nil {
		return local != "latest" && local != remote
	}
	if remoteSem.GreaterThan(localSem) {
		return true
	} else if remoteSem.Equal(localSem) {
		return remoteSem.Metadata() > localSem.Metadata()
	}
	return false
}
