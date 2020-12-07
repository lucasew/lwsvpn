# Run the actions in the container

if [ -n "$PORT" ]; then
    wstunnel --server ws://0.0.0.0:$PORT &
else
    echo "Missing PORT variable for wsvpn"
    exit 1
fi

if [ -n "$RCLONE_CFG" ]; then
    echo "$RCLONE_CFG" > rclone.conf
    rclone serve webdav --addr :9999 davsrv:/ --config ./rclone.conf &
fi
