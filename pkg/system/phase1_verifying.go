package system

import (
	"fmt"

	dockerref "github.com/docker/distribution/reference"
	semver "github.com/hashicorp/go-version"
	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"
)

// ReconcilePhaseVerifying runs the reconcile verify phase
func (r *Reconciler) ReconcilePhaseVerifying() error {

	r.SetPhase(
		nbv1.SystemPhaseVerifying,
		"SystemPhaseVerifying",
		"noobaa operator started phase 1/4 - \"Verifying\"",
	)

	if err := r.CheckSystemCR(); err != nil {
		return err
	}

	return nil
}

// CheckSystemCR checks the validity of the system CR
// (i.e system.metadata.name and system.spec.image)
// and updates the status accordingly
func (r *Reconciler) CheckSystemCR() error {

	log := r.Logger.WithField("func", "CheckSystemCR")

	// we assume a single system per ns here
	if r.NooBaa.Name != options.SystemName {
		return util.NewPersistentError("InvalidSystemName",
			fmt.Sprintf("Invalid system name %q expected %q", r.NooBaa.Name, options.SystemName))
	}

	specImage := options.ContainerImage
	if r.NooBaa.Spec.Image != nil {
		specImage = *r.NooBaa.Spec.Image
	}

	// Parse the image spec as a docker image url
	imageRef, err := dockerref.Parse(specImage)

	// If the image cannot be parsed log the incident and mark as persistent error
	// since we don't need to retry until the spec is updated.
	if err != nil {
		return util.NewPersistentError("InvalidImage",
			fmt.Sprintf(`Invalid image requested %q %v`, specImage, err))
	}

	// Get the image name and tag
	imageName := ""
	imageTag := ""
	switch image := imageRef.(type) {
	case dockerref.NamedTagged:
		log.Infof("Parsed image (NamedTagged) %v", image)
		imageName = image.Name()
		imageTag = image.Tag()
	case dockerref.Tagged:
		log.Infof("Parsed image (Tagged) %v", image)
		imageTag = image.Tag()
	case dockerref.Named:
		log.Infof("Parsed image (Named) %v", image)
		imageName = image.Name()
	default:
		log.Infof("Parsed image (unstructured) %v", image)
	}

	if imageName == options.ContainerImageName {
		version, err := semver.NewVersion(imageTag)
		if err == nil {
			log.Infof("Parsed version %q from image tag %q", version.String(), imageTag)
			if !ContainerImageConstraint.Check(version) {
				return util.NewPersistentError("InvalidImageVersion",
					fmt.Sprintf(`Invalid image version %q not matching constraints %q`,
						imageRef, ContainerImageConstraint))
			}
		} else {
			log.Infof("Using custom image %q constraints %q", imageRef.String(), ContainerImageConstraint.String())
		}
	} else {
		log.Infof("Using custom image name %q the default is %q", imageRef.String(), options.ContainerImageName)
	}

	// Set ActualImage to be updated in the noobaa status
	r.NooBaa.Status.ActualImage = specImage

	return nil
}
