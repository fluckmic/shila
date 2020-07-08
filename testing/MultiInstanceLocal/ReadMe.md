Before you run.sh, setup and start the local scion infrastructure using the 
provided Shila4.topo.

- Copy Shila4.topo into scionproto/scion/topology.
- Stop the current run (./scion.sh stop) of scion.
- Setup the topology (./scion.sh topology -c topology/Shila4.topo).
- Start a new run. (./scion.sh start)

Now you can run.sh.


LengthTest.topo
./scion.sh topology -c topology/LengthTest.topo 
1-ff00:0:112 sd:127.0.0.43:30255
2-ff00:0:220 sd:127.0.0.67:30255

MTUTest.topo
./scion.sh topology -c topology/MTUTest.topo
1-ff00:0:112 sd:127.0.0.19:30255
2-ff00:0:220 sd:127.0.0.27:30255

SharabilityTest.topo
./scion.sh topology -c topology/SharabilityTest.topo
1-ff00:0:112 sd:127.0.0.43:30255
2-ff00:0:220 sd:127.0.0.67:30255