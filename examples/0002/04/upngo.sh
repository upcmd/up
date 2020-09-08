#change to list and download linux version by yourself, this is only for demo

list_rolling(){
curl -s https://api.github.com/repos/upcmd/up/releases \
    |grep darwin_amd64_rolling \
    |grep download \
    |awk '{print $2}'  |cut -d \-  -f2|cut -d \/ -f1
}

download_rolling(){
if [ "$1" == "" ];then
    echo "syntax exaple: download_rolling 20200814"
else
    ver=$1
    curl -s https://api.github.com/repos/upcmd/up/releases \
        |grep darwin_amd64_rolling \
        |grep download \
        |grep $ver \
        |awk '{print $2}' \
        |xargs -I % curl -L % -o up \
        && chmod +x up
fi
}

#exaple:
#list_rolling
#download_rolling 20200814

#download the rolling release in case concerning the stability
#rolling_version=20200902
#download_rolling ${rolling_version}

#download the latest for quick test
curl -s https://api.github.com/repos/upcmd/up/releases \
    |grep darwin_amd64_latest \
    |grep download \
    |head -n 1 \
    |awk '{print $2}' \
    |xargs -I % curl -L % -o up \
    && chmod +x up

echo "eprofileid: $EProfileID"
./up ngo -p $EProfileID
