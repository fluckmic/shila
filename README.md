## Implementing and Evaluating MPTCP on the SCION Future Internet Architecture

This repository contains all relevant data of Michael A. Flückiger's master thesis (March to August 2020). 

The final report can be found [here]: report/Report-ImplAndEvalMPTCPonSCION-Maf.pdf	"Implementing and Evaluating MPTCP on the SCION Future Internet Architecture".

### Content

##### Implementation

##### Additional material

- *measurements* - Everything related to the measurements performed.

  - *performance* - Scripts for the Shila Measurement.
  - *quicT* - Scripts for the Quic Measurement.
  - *sessionScripts* - General helper scripts.
  - *post* - Scripts for post-processing the raw results (evaluation and plotting).

  

  

### Abstract

This work relies on the two technologies MPTCP and SCION. Multipath TCP (MPTCP) is an extension to the Transmission Control Protocol (TCP). In contrast to TCP uses MPTCP several connections, so-called flows, for data exchange. An approach that has recently become increasingly popular, fitting the needs of today's multihomed devices. SCION is a secure Internet architecture designed to address the weaknesses and shortcomings of today's Internet. It implements path transparency as an important feature. In contrast to the current Internet, SCION gives both the transmitter and receiver control and knowledge of the paths along which their data is exchanged.

In this thesis, we present the implementation and evaluation of Shila, an approach to combine these two technologies. With this name-giving shim layer, the use of TCP applications over the SCION network becomes possible. Thanks to Shila, the large number of such TCP applications can be tested via SCION without the need to change its implementation. If hosts support MPTCP, one also benefits from its advantages and the inherent support of multiple paths in SCION. For example, Shila allows the paths for the individual MPTCP flows to be selected according to different criteria, such as being as short as possible.

Our implementation uses virtual network interfaces for the interaction between Shila and applications. Created during startup offers each virtual interface the possibility for a single flow of an MPTCP connection. For data exchange between Shila instances on different hosts, backbone connections via the SCION network are set up once a TCP connection is about to happen. If a client binds to one of the virtual interfaces to establish a new TCP connection, the IP traffic is intercepted by Shila. The SCION address of the host running the server is determined using the TCP address extracted from the received datagram and a hardcoded mapping. Shila contacts its counterpart on the receiving side via a dedicated endpoint listening at this SCION address and a well-known port. A main-flow, holding a backbone connection for data exchange, is established and linked to the TCP connection. MPTCP now starts to initiate further flows via each additional available virtual interface. Linked with its main flow, Shila has all the information necessary to set up individual backbone connections for these sub-flows accordingly.

We have evaluated Shila in the SCIONLab network using iPerf3 as application. The measurement has shown, that the throughput can be increased by using multiple paths. Compared to the implementation of QUIC via SCION, our approach performs worse. The multiple detours through the kernel reduce the performance. Furthermore causes the sending of redundant header information via the backbone connection an unnecessarily high overhead. 

With the finally presented approaches to improve Shila, such as integrating its functionality into XDP to shorten the paths through the kernel, this work lays the foundation for continuing development, improvement and research, which will also benefit the further deployment of SCION.

### Shila

##### Setup

##### Usage
