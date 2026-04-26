/*
Copyright 2026 Pedro Cozinheiro.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package officer

import (
	"context"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
)

// ImageMetadataProvider allows resolving OCI image metadata.
type ImageMetadataProvider interface {
	// GetImageLabels returns OCI labels for a given image reference.
	GetImageLabels(ctx context.Context, image string) (map[string]string, error)
}

// resolveRepositoryURL extracts a repository URL from a Pod.
//
// Resolution order:
//  1. annotation: court.dev/repository
//  2. OCI label: org.opencontainers.image.source
//
// Returns empty string if nothing is found.
func resolveRepositoryURL(ctx context.Context, log logr.Logger, pod *v1.Pod, imgProvider ImageMetadataProvider) string {

	if pod == nil {
		return ""
	}

	// 1. explicit annotation (fast path)
	if repo, ok := pod.Annotations["court.dev/repository"]; ok && repo != "" {
		return repo
	}

	log.V(1).Info("repository not in annotations, checking image metadata")

	if imgProvider == nil {
		log.V(1).Info("image metadata provider not configured")
		return ""
	}

	// 2. OCI labels fallback
	for _, c := range pod.Spec.Containers {

		labels, err := imgProvider.GetImageLabels(ctx, c.Image)
		if err != nil {
			log.Error(err, "failed to inspect image", "image", c.Image)
			continue
		}

		// OCI standard label for source repository
		if repo, ok := labels["org.opencontainers.image.source"]; ok && repo != "" {
			return repo
		}
	}

	// No repository information found
	log.V(1).Info("repository not resolved from annotations or image metadata")

	return ""
}
