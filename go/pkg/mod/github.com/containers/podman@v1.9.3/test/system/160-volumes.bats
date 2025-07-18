#!/usr/bin/env bats   -*- bats -*-
#
# podman volume-related tests
#

load helpers

function setup() {
    basic_setup

    run_podman '?' volume rm -a
}

function teardown() {
    run_podman '?' rm -a --volumes
    run_podman '?' volume rm -a -f

    basic_teardown
}


# Simple volume tests: share files between host and container
@test "podman run --volumes : basic" {
    skip_if_remote "volumes cannot be shared across hosts"

    # Create three temporary directories
    vol1=${PODMAN_TMPDIR}/v1_$(random_string)
    vol2=${PODMAN_TMPDIR}/v2_$(random_string)
    vol3=${PODMAN_TMPDIR}/v3_$(random_string)
    mkdir $vol1 $vol2 $vol3

    # In each directory, write a random string to a file
    echo $(random_string) >$vol1/file1_in
    echo $(random_string) >$vol2/file2_in
    echo $(random_string) >$vol3/file3_in

    # Run 'cat' on each file, and compare against local files. Mix -v / --volume
    # flags, and specify them out of order just for grins. The shell wildcard
    # expansion must sort vol1/2/3 lexically regardless.
    v_opts="-v $vol1:/vol1:z --volume $vol3:/vol3:z -v $vol2:/vol2:z"
    run_podman run --rm $v_opts $IMAGE sh -c "cat /vol?/file?_in"

    for i in 1 2 3; do
        eval voldir=\$vol${i}
        is "${lines[$(($i - 1))]}" "$(< $voldir/file${i}_in)" \
           "contents of /vol${i}/file${i}_in"
    done

    # Confirm that container sees vol1 as a mount point
    run_podman run --rm $v_opts $IMAGE mount
    is "$output" ".* on /vol1 type .*" "'mount' in container lists vol1"

    # Have the container do write operations, confirm them on host
    out1=$(random_string)
    run_podman run --rm $v_opts $IMAGE sh -c "echo $out1 >/vol1/file1_out;
                                              cp /vol2/file2_in /vol3/file3_out"
    is "$(<$vol1/file1_out)" "$out1"              "contents of /vol1/file1_out"
    is "$(<$vol3/file3_out)" "$(<$vol2/file2_in)" "contents of /vol3/file3_out"

    # Writing to read-only volumes: not allowed
    run_podman 1 run --rm -v $vol1:/vol1ro:z,ro $IMAGE sh -c "touch /vol1ro/abc"
    is "$output" ".*Read-only file system"  "touch on read-only volume"
}


# Running scripts (executables) from a volume
@test "podman volume: exec/noexec" {
    myvolume=myvol$(random_string)

    run_podman volume create $myvolume
    is "$output" "$myvolume" "output from volume create"

    run_podman volume inspect --format '{{.Mountpoint}}' $myvolume
    mountpoint="$output"

    # Create a script, make it runnable
    rand=$(random_string)
    cat >$mountpoint/myscript <<EOF
#!/bin/sh
echo "got here -$rand-"
EOF
    chmod 755 $mountpoint/myscript

    # By default, volumes are mounted exec, but we have manually added the
    # noexec option. This should fail.
    # ARGH. Unfortunately, runc (used for cgroups v1) produces a different error
    local expect_rc=126
    local expect_msg='.* OCI runtime permission denied.*'
    run_podman info --format '{{ .Host.OCIRuntime.Path }}'
    if expr "$output" : ".*/runc"; then
        expect_rc=1
        expect_msg='.* exec user process caused.*permission denied'
    fi

    run_podman ${expect_rc} run --rm --volume $myvolume:/vol:noexec,z $IMAGE /vol/myscript
    is "$output" "$expect_msg" "run on volume, noexec"

    # With the default, it should pass
    run_podman run --rm -v $myvolume:/vol:z $IMAGE /vol/myscript
    is "$output" "got here -$rand-" "script in volume is runnable with default (exec)"

    # Clean up
    run_podman volume rm $myvolume
}


# Confirm that container sees the correct id
@test "podman volume with --userns=keep-id" {
    is_rootless || skip "only meaningful when run rootless"

    myvoldir=${PODMAN_TMPDIR}/volume_$(random_string)
    mkdir $myvoldir
    touch $myvoldir/myfile

    # With keep-id
    run_podman run --rm -v $myvoldir:/vol:z --userns=keep-id $IMAGE \
               stat -c "%u:%s" /vol/myfile
    is "$output" "$(id -u):0" "with keep-id: stat(file in container) == my uid"

    # Without
    run_podman run --rm -v $myvoldir:/vol:z $IMAGE \
               stat -c "%u:%s" /vol/myfile
    is "$output" "0:0" "w/o keep-id: stat(file in container) == root"
}


# 'volume prune' identifies and cleans up unused volumes
@test "podman volume prune" {
    # Create four named volumes
    local -a v=()
    for i in 1 2 3 4;do
        vol=myvol${i}$(random_string)
        v[$i]=$vol
        run_podman volume create $vol
    done

    # Run two containers: one mounting v1, one mounting v2 & v3
    run_podman run --name c1 --volume ${v[1]}:/vol1 $IMAGE date
    run_podman run --name c2 --volume ${v[2]}:/vol2 -v ${v[3]}:/vol3 \
               $IMAGE date

    # prune should remove v4
    run_podman volume prune --force
    is "$output" "${v[4]}" "volume prune, with 1, 2, 3 in use, deletes only 4"

    # Remove the container using v2 and v3. Prune should now remove those.
    # The 'echo sort' is to get the output sorted and in one line.
    run_podman rm c2
    run_podman volume prune --force
    is "$(echo $(sort <<<$output))" "${v[2]} ${v[3]}" \
       "volume prune, after rm c2, deletes volumes 2 and 3"

    # Remove the final container. Prune should now remove v1.
    run_podman rm c1
    run_podman volume prune --force
    is "$output"  "${v[1]}" "volume prune, after rm c2 & c1, deletes volume 1"

    # Further prunes are NOPs
    run_podman volume prune --force
    is "$output"  "" "no more volumes to prune"
}


# vim: filetype=sh
