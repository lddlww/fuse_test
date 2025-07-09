//go:build linux || freebsd

package integration

var (
	STORAGE_FS          = "overlay"
	STORAGE_OPTIONS     = "--storage-driver overlay"
	ROOTLESS_STORAGE_FS = "overlay"
	CACHE_IMAGES        = []string{ALPINE, BB, fedoraMinimal, NGINX_IMAGE, REDIS_IMAGE, REGISTRY_IMAGE, INFRA_IMAGE, CITEST_IMAGE, HEALTHCHECK_IMAGE, SYSTEMD_IMAGE}
	NGINX_IMAGE         = "quay.io/lsm5/alpine_nginx-aarch64:latest"
	BB_GLIBC            = "docker.io/library/busybox:glibc"
	REGISTRY_IMAGE      = "quay.io/libpod/registry:2.8.2"
	CITEST_IMAGE        = "quay.io/libpod/testimage:20241011"
	SYSTEMD_IMAGE       = "quay.io/libpod/systemd-image:20240124"
	CIRROS_IMAGE        = "quay.io/libpod/cirros:latest"
)
