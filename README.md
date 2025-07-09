# rclone+MinIO VS JuiceFS+MinIO Performance Comparison (Small Files Test)

This repository demonstrates a performance comparison between **rclone** and **JuiceFS**, both mounted via FUSE, when working with **many small files** stored in the same **MinIO** backend.

---

## Components Used

- [MinIO](https://min.io/) — object storage backend
- [rclone](https://rclone.org/) — used for S3 mount with FUSE
- [JuiceFS](https://juicefs.com/) — POSIX-compatible file system
- Redis — used as JuiceFS metadata engine

---

## Deployment & Testing Steps

## Step 1: Ensure MinIO is running
#### - Endpoint: http://192.168.10.103:9000
#### - Access Key: root
#### - Secret Key: xxxx
#### - Create two buckets: bucket1 (for rclone) and jfs (for JuiceFS)

## Step 2: Install rclone
```bash
wget https://github.com/rclone/rclone/releases/download/v1.70.2/rclone-v1.70.2-linux-amd64.deb
sudo dpkg -i rclone-v1.70.2-linux-amd64.deb
```
## Step 3: Install JuiceFS
```bash
curl -sSL https://d.juicefs.com/install | sh -
```
## Step 4: Configure rclone with MinIO
```bash
rclone config create myminio s3 \
  provider=Minio \
  access_key_id=root \
  secret_access_key=xxxx \
  endpoint=http://192.168.10.103:9000 \
  region=cn \
  acl=private
```

## Step 5: Format JuiceFS
```bash
export METAURL=redis://:Mj43eU6PRtADDuf8@192.168.10.103:6379/1
juicefs format \
  --storage minio \
  --bucket http://192.168.10.103:9000/jfs \
  --access-key root \
  --secret-key xxxx \
  --capacity 100 \
  $METAURL jfs
```

## Step 6: Mount rclone via FUSE
```bash
mkdir -pv /data/rclone
rclone mount myminio:bucket1 /data/rclone \
  --vfs-cache-mode full \
  --dir-cache-time=72h \
  --vfs-cache-max-age 24h \
  --vfs-cache-max-size 100G \
  --transfers 32 \
  --checkers 64 \
  --write-back-cache \
  --attr-timeout 1h \
  --no-checksum \
  --no-modtime \
  --allow-other \
  --daemon
```

## Step 7: Mount JuiceFS via FUSE
```bash
mkdir -pv /data/juicefs
juicefs mount $METAURL /data/juicefs -d --writeback
```

## Step 8: Run small file copy tests
```bash
time cp -a fuse_test/go /data/rclone

time cp -a fuse_test/go /data/juicefs
```

