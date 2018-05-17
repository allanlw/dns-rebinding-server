FROM ubuntu:xenial
MAINTAINER Danny Cooper <dcooper@akamai.com>

RUN \
    apt-get update && \
    apt-get -y install \
        jq \
        tcpdump \
        curl \
        psmisc \
        wget

RUN \
wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - && \
echo 'deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main' | tee /etc/apt/sources.list.d/google-chrome.list

RUN \
    apt-get update && \
    apt-get -y install \
        firefox \
        chromium-browser \
        google-chrome-stable



ADD ./tests /tests
WORKDIR /tests
RUN mkdir ./captures
ENTRYPOINT ["sh", "start_all.sh"]

