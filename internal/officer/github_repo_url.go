package officer

import (
	"context"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
)

// ImageMetadataProvider defines the contract required to fetch
// metadata (labels) from a container image.
type ImageMetadataProvider interface {
	// GetImageLabels returns OCI labels for a given image reference.
	GetImageLabels(ctx context.Context, image string) (map[string]string, error)
}

// resolveRepositoryURL attempts to extract a GitHub repository URL
// from a given Pod.
//
// Resolution order:
//  1. court.dev/repository annotation
//  2. OCI image label: org.opencontainers.image.source
//
// If no repository can be resolved, an empty string is returned
// and a warning is logged.
func resolveRepositoryURL(ctx context.Context, log logr.Logger, pod *v1.Pod, imgProvider ImageMetadataProvider) string {

	if pod == nil {
		return ""
	}

	// 1. Explicit annotation
	if repo, ok := pod.Annotations["court.dev/repository"]; ok && repo != "" {
		return repo
	}

	log.Info("repository annotation missing, attempting OCI label resolution",
		"annotation", "court.dev/repository",
	)

	// 2. OCI label lookup via container runtime
	if imgProvider == nil {
		log.Info("image metadata provider not configured, cannot resolve repository")
		return ""
	}

	for _, c := range pod.Spec.Containers {

		labels, err := imgProvider.GetImageLabels(ctx, c.Image)
		if err != nil {
			log.Error(err, "failed to fetch image labels",
				"image", c.Image,
			)
			continue
		}

		// OCI standard label for source repository
		if repo, ok := labels["org.opencontainers.image.source"]; ok && repo != "" {
			return repo
		}
	}

	// No repository information found
	log.Info("unable to resolve repository from annotations or OCI labels")

	return ""
}
