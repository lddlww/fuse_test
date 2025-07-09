package integration

var (
	STORAGE_FS               = "vfs"
	STORAGE_OPTIONS          = "--storage-driver vfs"
	ROOTLESS_STORAGE_FS      = "vfs"
	ROOTLESS_STORAGE_OPTIONS = "--storage-driver vfs"
	CACHE_IMAGES             = []string{ALPINE, BB, fedoraMinimal, nginx, redis, registry, infra, labels, healthcheck}
	nginx                    = "quay.io/libpod/alpine_nginx:latest"
	BB_GLIBC                 = "docker.io/library/busybox:glibc"
	registry                 = "docker.io/library/registry:2"
	labels                   = "quay.io/libpod/alpine_labels:latest"
)
