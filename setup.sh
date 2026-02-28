#!/usr/bin/env bash
curl -L https://github.com/erebe/wstunnel/releases/download/v5.0/wstunnel-linux-x64 -o wstunnel
chmod +x wstunnel
curl -L https://downloads.rclone.org/v1.54.0/rclone-v1.54.0-linux-amd64.zip -o rclone.zip
unzip rclone.zip
mv rclone-*-linux-amd64/rclone ./rclone
