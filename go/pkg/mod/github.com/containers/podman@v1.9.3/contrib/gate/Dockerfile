FROM fedora:31
RUN dnf -y install \
      btrfs-progs-devel \
      bzip2 \
      container-selinux \
      containernetworking-cni \
      device-mapper-devel \
      findutils \
      git \
      glib2-devel \
      glibc-static \
      golang \
      gpgme-devel \
      iptables \
      libassuan-devel \
      libseccomp-devel \
      libselinux-devel \
      lsof \
      make \
      nmap-ncat \
      procps-ng \
      python \
      python3-dateutil \
      python3-pip \
      python3-psutil \
      python3-pytoml \
      python3-pyyaml \
      python3-varlink \
      rsync \
      slirp4netns \
      unzip \
      which \
      xz \
      zip \
      && dnf clean all

ENV GOPATH="/var/tmp/go" \
    GOBIN="/var/tmp/go/bin" \
    PATH="/var/tmp/go/bin:/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin" \
    SRCPATH="/usr/src/libpod" \
    GOSRC="/var/tmp/go/src/github.com/containers/libpod"

# Only needed for installing build-time dependencies, then will be removed
COPY / $GOSRC

# Install dependencies
RUN set -x && \
    mkdir -p "$GOBIN" && \
    mkdir -p /etc/cni/net.d && \
    mkdir -p /etc/containers && \
    install -D -m 755 $GOSRC/contrib/gate/entrypoint.sh /usr/local/bin/ && \
    python3 -m pip install pre-commit

# Install cni config
COPY cni/87-podman-bridge.conflist /etc/cni/net.d/87-podman-bridge.conflist
# Make sure we have some policy for pulling images
COPY test/policy.json /etc/containers/policy.json
COPY test/redhat_sigstore.yaml /etc/containers/registries.d/registry.access.redhat.com.yaml

WORKDIR "$GOSRC"
RUN make install.tools && \
    cd / && \
    rm -rf "$GOSRC" && \
    mkdir -p "$GOSRC"
VOLUME ["/usr/src/libpod"]
# This entrypoint will synchronize the above volume ($SRCPATH) to $GOSRC before
# executing make.  This ensures the original source remains prestine and is never
# modified by any lint/validation checks.
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
