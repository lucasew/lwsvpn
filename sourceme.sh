export TAG=wsvpn
export PORT=1025

container_run() {
    docker run -ti --env PORT=$PORT --env "RCLONE_CFG=$(cat ~/.config/rclone/rclone.conf)" -p 3000:$PORT $TAG
}

container_build() {
    docker build . -t $TAG
}
