wget -O /tmp/wstunnel.zip https://github.com/erebe/wstunnel/releases/download/v3.0/wstunnel-x64-linux.zip
unzip /tmp/wstunnel.zip -d /tmp
install -m755 /tmp/wstunnel /bin

rm /tmp/* -rf

