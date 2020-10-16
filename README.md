# SCHRODINGER Final Project
Final Project for CSCI 6421 

Team Members
```
Xuzheng Lu (Leader)
Zetian Zheng
Xiaoyu Shen
Guangyuan Shen
```
## Topic: Distributed Hash Table

A Distributed Hash Table (DHT) is a distributed system that provides a lookup service similar to a hash table. The network maintains a huge file index hash table, divided and stored on each node of the network according to certain rules, with entries in the form of (key, value). Usually, the key is the hash value of the file, and the value is the IP address where the file is stored. Given the key, the value/address stored in the node can be efficiently found and returned to the query node.

[(Wikipedia: Distributed hash table)](https://en.wikipedia.org/wiki/Distributed_hash_table?oldformat=true)

<p align="center">
  <img width="600" src="https://upload.wikimedia.org/wikipedia/commons/thumb/9/98/DHT_en.svg/1000px-DHT_en.svg.png" />
</p>

**State-of-the-art DHT algorithms (for reference):**
1. **Chord** - a protocol and algorithm for a peer-to-peer distributed hash table
[(Wikipedia: Chord (peer-to-peer))](https://en.wikipedia.org/wiki/Chord_(peer-to-peer)?oldformat=true)

2. **Pastry** - an overlay network and routing network for the implementation of a distributed hash table similar to Chord.
[(Wikipedia: Pastry (DHT))](https://en.wikipedia.org/wiki/Pastry_(DHT)?oldformat=true)

3. **Tapestry** - a peer-to-peer overlay network which provides a distributed hash table, routing, and multicasting infrastructure for distributed applications.
[(Wikipedia: Tapestry (DHT))](https://en.wikipedia.org/wiki/Tapestry_(DHT)?oldformat=true)

## Challenges
1. **Autonomy and decentralization**
- In this DHT system, the nodes should be autonomous and decentralized, which means it has no central coordination.
- For each object, the node responsible for the object should be reachable through a short routing path.
- The number of neighbors of each node should be kept reasonable.
2. **Fault tolerance**
- The DHT system should be reliable and robust in any case, and avoid the crash from high concurrency.
- In this project, we can use algorithmic controlling of the distributed system’s components to provide the desired service (Storm C, 2012).

3. **Scalability**
- The DHT should be able to handle a growing amount of nodes be added to the current system.
- The system should be able to handle the node addition and removal:
  - repartition the affected keys on the existing node;
  - reorganize the neighbor nodes;
  - to connect new nodes to DHT through a guiding mechanism.
- For example, in Apache Cassandra (John Hammink, 2019), an open-source NoSQL DHT system, each node can communicate with a constant amount of other nodes, which let the system to scale linearly over a huge number of nodes. 

## References
1. Byers, John, Jeffrey Considine, and Michael Mitzenmacher. "Simple load balancing for distributed hash tables." International Workshop on Peer-to-Peer Systems. Springer, Berlin, Heidelberg, 2003.
2. Stoica, Ion, et al. "Chord: a scalable peer-to-peer lookup protocol for internet applications." IEEE/ACM Transactions on networking 11.1 (2003): 17-32.
3. Talia, Domenico, and Paolo Trunfio. "Enabling dynamic querying over distributed hash tables." Journal of Parallel and Distributed Computing 70.12 (2010): 1254-1265.
4. Dougherty, Michael, Haklin Kimm, and Ho-sang Ham. "Implementation of the distributed hash tables on peer-to-peer networks." 2008 IEEE Sarnoff Symposium. IEEE, 2008.
5. Awerbuch, Baruch, and Christian Scheideler. "Towards a scalable and robust DHT." Theory of Computing Systems 45.2 (2009): 234-260.
6. Storm, Christian. Specification and Analytical Evaluation of Heterogeneous Dynamic Quorum-Based Data Replication Schemes. Springer Science & Business Media, 2012.
7. [John Hammink. "An introduction to Apache Cassandra". 2019.](https://aiven.io/blog/an-introduction-to-apache-cassandra#:~:text=This%20is%20one%20of%20the,and%20data%20centers%20go%20down)
