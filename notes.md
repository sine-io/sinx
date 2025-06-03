
scoop bucket add extras
scoop install extras/protobuf

git config --global http.proxy <http://ip:port>
git config --global https.proxy <http://ip:port>

git config --global --unset http.proxy
git config --global --unset https.proxy

git config --global --get http.proxy
git config --global --get https.proxy
