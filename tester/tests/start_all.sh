#!/bin/sh
querystring=$1

base_domain="rebindmy.zone"
poll_domain="$base_domain/record"
attack_domain="$base_domain/attack"
base_uid=`date +%s | tail -c4`

capture_dir=/results/captures/
log_dir=/results/logs/
mkdir -p $capture_dir
mkdir -p $log_dir

firefox_uid='firefox'$base_uid
chromium_uid='chromium'$base_uid
googlechrome_uid='gchrome'$base_uid


firefox_querystring="$querystring&id=$firefox_uid"
tcpdump -i eth0 -s 65535 -w $capture_dir/$firefox_uid.pcap &
firefox -headless "$attack_domain$firefox_querystring" &
firefox_status=''
until [ "$firefox_status" = '"done"' ]
do
    firefox_status=`curl $poll_domain"?id=$firefox_uid" | jq .status?`
    echo $firefox_status
    sleep 10s
done
mkdir -p $log_dir/$firefox_uid
curl $poll_domain"?id=$firefox_uid" > $log_dir/$firefox_uid/final_response.json
killall tcpdump


chromium_querystring="$querystring&id=$chromium_uid"
tcpdump -i eth0 -s 65535 -w $capture_dir/$chromium_uid.pcap &
chromium-browser --headless --disable-gpu --remote-debugging-port=9999 --no-sandbox "http://$attack_domain$chromium_querystring" &
chromium_status=''
until [ "$chromium_status" = '"done"' ]
do
    chromium_status=`curl $poll_domain"?id=$chromium_uid" | jq .status?`
    echo $chromium_status
    sleep 10s
done
mkdir -p $log_dir/$chromium_uid
curl $poll_domain"?id=$chromium_uid" > $log_dir/$chromium_uid/final_response.json
killall tcpdump

gchrome_querystring="$querystring&id=$googlechrome_uid"
tcpdump -i eth0 -s 65535 -w $capture_dir/$googlechrome_uid.pcap &
google-chrome --disable-gpu --headless --no-sandbox --remote-debugging-port=9999 "http://$attack_domain$gchrome_querystring"
googlechrome_status=''
until [ $googlechrome_status = '"done"' ]
do
    $googlechrome_status=`curl $poll_domain"?id=$googlechrome_uid" | jq .status?`
    echo $googlechrome_status
    sleep 10s
done
mkdir -p $log_dir/$googlechrome_uid
curl $poll_domain"?id=$googlechrome_uid" > $log_dir/$googlechrome_uid/final_response.json
killall tcpdump



