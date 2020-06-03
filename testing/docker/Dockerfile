FROM ubuntu:16.04

# Install some essentials
RUN 	apt-get update
RUN	apt-get install -y git-core build-essential iperf wget

# Install go
RUN 	wget https://dl.google.com/go/go1.14.4.linux-amd64.tar.gz 
RUN	tar -C /usr/local/ -xzf go1.14.4.linux-amd64.tar.gz
RUN	rm go1.14.4.linux-amd64.tar.gz

ENV	GOROOT=/usr/local/go
ENV	GOPATH=$HOME
ENV	PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Fetch shila
RUN 	mkdir -p /root/go/src/github.com/google
WORKDIR /root/go/src
RUN 	git clone https://github.com/fluckmic/shila.git
WORKDIR /root/go/src/github.com/google
RUN 	git clone https://github.com/google/gopacket.git
