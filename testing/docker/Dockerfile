FROM ubuntu:16.04

# Install some essentials
RUN 	apt-get update
RUN	apt-get install -y git-core build-essential apt-transport-https iperf wget pkg-config gnupg-curl

# Get MPTCP tools
RUN 	apt-key adv --keyserver hkps://keyserver.ubuntu.com:443 --recv-keys 379CE192D401AB61
RUN 	sh -c "echo 'deb https://dl.bintray.com/multipath-tcp/mptcp_deb stable main' > /etc/apt/sources.list.d/mptcp.list"
RUN 	apt-get update -y
RUN	apt-get dist-upgrade -y

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
WORKDIR /root/go/src/shila
