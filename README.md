## Implementing and Evaluating MPTCP on the SCION Future Internet Architecture

This repository contains all relevant data of Michael A. Flückiger's master thesis (March to August 2020). 

The final report can be found [here](https://github.com/fluckmic/shila/blob/hand-in/report/Report-ImplAndEvalMPTCPonSCION-Maf.pdf).

### Content

(Note: The **bold** words correspond to terms that are used in the report. )

##### Implementation

- *config/config.go* - Holds the default configuration of Shila.
- *core/*
  - *connection/* - Contains the implementation of the **Shila-Connection**.
  - *router/* - Contains the implementation of the **Router** and the **Path Selection** functionality.
  - *shila/* - Contains implementation of core concepts like the **Shila-Packet**, the **TCP-Flow** and the **Net-Flow**.
- *io/structure/* 
  - *config.go* - Defines and describes all configurable parameters.
- *kernelSide/*
  - *kernelSide.go* - Implementation of the **Kernel-Side**.
  - *kernelEndpoint/* - Contains the implementation of the **Kernel-Endpoint**.
- *layer/* - Contains the parsing functionality required to parse the IP and the TCP/MPTCP packets.
- *log/logging.go* - Contains the implementation of the logging functionality.
- *networkSide/* 
  - *networkEndpoint/* - Contains the implementation of the **Client-Endpoint**, the **Server-Endpoint** and the **Backbone-Connection**.
  - *networkSide.go* - Implementation of the **Network-Side** (supplemented by networkSideSpecific.go).
- *shutdown/shutdown.go* - Implementation of the termination functionality.
- *workingSide/* - Implementation of the **Working-Side**.
- *config.json* - Optional file holding configuration parameters specified by the user.
- *main.go* - Entry point of Shila.

##### Additional material

- *report/* - The final report and all the raw material and drawings.
- *presentation/* - The final presentation and all its raw content.

- *testing/* - Scripts facilitating testing, mainly by setting up the difference instances.
  - *local/* - For local testing.
  - *SCIONLab/* - For testing withing the SCIONLab network using the custom ASes.

- *measurements/* - Everything related to the measurements performed.
  - *performance/* - Scripts for the Shila Measurement.
  - *quicT/* - Scripts for the Quic Measurement.
  - *sessionScripts/* - General helper scripts.
  - *post/* - Scripts for post-processing the raw results (evaluation and plotting).
  - *results/* - The raw unprocessed data gathered in the measurements.
- *helper/netnsClear.sh* - Helper script to clean up networks namespaces used during debugging.

### Abstract

This work relies on the two technologies MPTCP and SCION. Multipath TCP (MPTCP) is an extension to the Transmission Control Protocol (TCP). In contrast to TCP, MPTCP uses several sub-connections, so-called flows, for data exchange. An approach that has recently become increasingly popular, fitting the needs of today's multihomed devices. SCION is a secure internet architecture designed to address the weaknesses and shortcomings of today's Internet. It implements path transparency as an important feature. In contrast to the current Internet, SCION gives both, the sender and the receiver, control and knowledge of the paths along where their data is exchanged.

In this thesis, we present the implementation and evaluation of Shila, an approach to combine these two technologies. With this name-giving shim layer, the use of TCP applications over the SCION network becomes possible. Thanks to Shila, the large number of such TCP applications can be tested via SCION without the need to change its implementation. If hosts support MPTCP, one also benefits from its advantages and the inherent support of multiple paths in SCION. For example, Shila allows the paths for the individual MPTCP flows to be selected according to different criteria, such as being as short as possible.

Our implementation uses virtual network interfaces for the interaction between Shila and the applications. Created during startup of Shila, each virtual interface offers the possibility for a single flow to an MPTCP connection. For data exchange between Shila instances on different hosts, backbone connections via the SCION network are set up once a TCP connection is about to happen. If a client binds to one of the virtual interfaces to establish a new TCP connection, the IP traffic is intercepted by Shila. The SCION address of the host running the server is determined using the TCP address extracted from the received datagram and a hardcoded mapping. Shila contacts its counterpart on the receiving side via a dedicated endpoint listening at this SCION address and a well-known port. A main-flow, holding a backbone connection for data exchange, is established and linked to the TCP connection. MPTCP now starts to initiate further flows via each additional available virtual interface. Linked with its main-flow, Shila has all the information necessary to set up individual backbone connections for these sub-flows accordingly.

We have evaluated Shila in the SCIONLab network using iPerf3 as an exemplary application. The measurement has shown, that the throughput can be increased by using multiple paths. Compared to the implementation of QUIC via SCION, our approach performs worse. The detour through the kernel and Shila reduces the performance. Furthermore, the sending of redundant header information via the backbone connections causes an unnecessarily high overhead.

With the finally presented approaches to improve Shila this work lays the foundation for continuing development, improvement and research, which will also benefit the further deployment of SCION.

### Shila

In the following we describe how to setup the environment to be able to use Shila and how to use it once setup is done.

##### Setup

###### Multipath TCP (MPTCP)

Install MPTCP as described [here](http://multipath-tcp.org/pmwiki.php/Users/AptRepository).

Install the MPTCP iproute-extension as described [here](http://multipath-tcp.org/pmwiki.php/Users/Tools) on the very top of the page.

Enable and configure MPTCP as described [here](http://multipath-tcp.org/pmwiki.php/Users/Tools). A good starting point is to use the *fullmesh* path-manager and the *default* scheduler.

###### Install SCION

Install SCION as described [here](https://github.com/netsec-ethz/scion), this includes the installation of Go.

After the installation of SCION you should have a folder *~/go/src/github.com/scionproto/scion/* containing the installation.

###### Install Shila

Get Shila and fetch the required packages.

`cd ~/go/src/github.com`

`git clone https://github.com/fluckmic/shila.git`

`cd shila`

`go get`

###### Install iPerf3 

`sudo apt install iperf3`

This is just needed for the subsequent described example usage.

##### Usage

Good stating point is to use Shila with a local setup of SCION and through prepared scripts in *shila/testing/local*.



##### Configuration

There are many parameters which can be used to configure Shila. This parameters are described [here](https://github.com/fluckmic/shila/blob/hand-in/io/structure/config.go), whereas the default values are set [here](https://github.com/fluckmic/shila/blob/hand-in/config/config.go).